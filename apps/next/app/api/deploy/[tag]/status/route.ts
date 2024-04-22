import { clerkClient } from "@clerk/nextjs";
import { NextResponse, type NextRequest } from "next/server";

export async function GET(
  request: NextRequest,
  { params: { tag } }: { params: { tag: string } },
) {
  const authStatus = await clerkClient.authenticateRequest({ request });
  if (!authStatus.isSignedIn) {
    return new Response(JSON.stringify({}), {
      status: 401,
    });
  }

  return new NextResponse(
    JSON.stringify({
      messsage: 'Tag found',
      tag: tag,
    }),
    {
      status: 200,
      headers: { "Content-Type": "application/json" },
    },
  )
}
