import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { spawnSync } from "node:child_process";
import { createWriteStream, existsSync, mkdirSync } from "node:fs";
import { Writable } from "node:stream";

export const dynamic = "force-dynamic";

export async function POST(request: NextRequest) {
  const authStatus = await clerkClient.authenticateRequest({ request });
  if (!authStatus.isSignedIn) {
    return new Response(JSON.stringify({}), {
      status: 401,
    });
  }

  if (!request.body) {
    return new Response(JSON.stringify({}), { status: 400 });
  }

  const tempDir = 'temp';
  if (!existsSync(tempDir)) {
    mkdirSync(tempDir);
  }

  const tag = `foo-${Date.now()}`;
  const filename = `${tag}.gz`;
  const uploadedTarball = Writable.toWeb(
    createWriteStream(filename, "binary")
  ) as WritableStream<Uint8Array>;
  await request.body.pipeTo(uploadedTarball);
  spawnSync("tar", ["xzfv", filename, "-C", "temp/"]);

  return new Response(
    JSON.stringify({ message: "Deployment Started", tag }),
    { status: 200 }
  );
}
