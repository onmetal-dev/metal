import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { createWriteStream, existsSync, mkdirSync } from "node:fs";
import { extract } from "tar";

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
  const uploadedTarball = createWriteStream(filename, "utf8");
  const extractor = extract({
    C: tempDir,
  });

  const writableUploadStream = new WritableStream<Uint8Array>({
    write(chunk) {
      uploadedTarball.write(chunk);
      extractor.write(Buffer.from(chunk));
    },
  });

  await request.body?.pipeTo(writableUploadStream);
  uploadedTarball.close();

  return new Response(
    JSON.stringify({ message: "Deployment Started", tag }),
    { status: 200 }
  );
}
