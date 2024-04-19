import { clerkClient } from "@clerk/nextjs";
import { createWriteStream } from "node:fs";

export const dynamic = "force-dynamic";

export async function POST(request: Request) {
  const authStatus = await clerkClient.authenticateRequest({ request });
  if (!authStatus.isSignedIn) {
    return new Response(JSON.stringify({}), {
      status: 401,
    });
  }

  const tag = `foo-${Date.now()}`;
  const uploadedTarball = createWriteStream(`${tag}.gz`, "utf8");
  const writableUploadStream = new WritableStream<Uint8Array>({
    write(chunk) {
      uploadedTarball.write(chunk);
    },
  });
  await request.body?.pipeTo(writableUploadStream);

  return new Response(JSON.stringify({ message: "Deployment Started", tag }, null, 2), {
    status: 200,
  });
}
