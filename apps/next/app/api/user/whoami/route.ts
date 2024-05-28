import { db } from "@/app/server/db";
import { users } from "@/app/server/db/schema";
import { decodeJwt } from "@clerk/backend/jwt";
import { clerkClient } from "@clerk/clerk-sdk-node";
import { eq } from "drizzle-orm";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const authStatus = await clerkClient.authenticateRequest(request);
  if (!authStatus.isSignedIn) {
    return new Response(JSON.stringify({}), {
      status: 401,
    });
  }
  const { payload: token } = decodeJwt(authStatus.token);
  const clerkUserId = token.sub;
  try {
    const user = await clerkClient.users.getUser(clerkUserId);
  } catch (error: any) {
    if (error.clerkError && error.errors.length > 0) {
      // handle resource_not_found (i.e. clerk user deleted)
      if (error.errors[0].code === "resource_not_found") {
        return new Response(JSON.stringify({}), {
          status: 401,
        });
      }
      console.error(JSON.stringify(error.errors[0]));
    }
    throw error;
  }

  // response is the user object in our db along with the token for api access
  const user = await db
    .select()
    .from(users)
    .where(eq(users.clerkId, clerkUserId))
    .limit(1)
    .then((rows) => rows[0] || null);
  if (!user) {
    return new Response(JSON.stringify({}), {
      status: 404,
    });
  }
  return new Response(
    JSON.stringify({ token: authStatus.token, user }, null, 2),
    {
      status: 200,
    }
  );
}
