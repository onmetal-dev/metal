import { clerkClient } from "@clerk/nextjs";
import { decodeJwt } from "@clerk/nextjs/server";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const authStatus = await clerkClient.authenticateRequest({ request });
  if (!authStatus.isSignedIn) {
    return new Response(JSON.stringify({}), {
      status: 401,
    });
  }
  const { payload } = decodeJwt(authStatus.token);
  const userId = payload.sub;
  const user = await clerkClient.users.getUser(userId);

  return new Response(JSON.stringify({ token: payload, user }, null, 2), {
    status: 200,
  });
}
