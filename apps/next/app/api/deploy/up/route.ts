import { getDirectoryNameForUps } from "@/app/server/util/functions";
import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { spawn } from "node:child_process";
import { createWriteStream, existsSync, mkdirSync } from "node:fs";
import { Writable } from "node:stream";
import { dirSync } from "tmp";

export const dynamic = "force-dynamic";
const upsDirectory = getDirectoryNameForUps();

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

  const tag = `up_${Date.now()}`;
  if (!existsSync(upsDirectory)) {
    mkdirSync(upsDirectory);
  }
  const { name: tempDirName } = dirSync({
    tmpdir: upsDirectory,
    name: tag,
  });

  const filename = `${tag}.gz`;
  const uploadedTarball = Writable.toWeb(
    createWriteStream(filename, "binary")
  ) as WritableStream<Uint8Array>;
  await request.body.pipeTo(uploadedTarball);

  const extractionStream = spawn('tar', ['xzfv', filename, '-C', tempDirName]);
  await new Promise<void>((resolve, reject) => {
    extractionStream.on('exit', () => {
      console.log("Tarball extracted");
      resolve();
    });
    extractionStream.on('error', (error) => {
      reject(error);
    })
  });

  return new Response(
    JSON.stringify({ message: "Deployment Started", tag }),
    { status: 200 }
  );
}
