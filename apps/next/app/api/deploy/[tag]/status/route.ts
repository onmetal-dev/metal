import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { exec as execCallbackBased } from "node:child_process";
import { promisify } from "node:util";

const exec = promisify(execCallbackBased);

// This code below is heavily based on:
// https://nextjs.org/docs/app/building-your-application/routing/route-handlers#streaming

// https://developer.mozilla.org/docs/Web/API/ReadableStream#convert_async_iterator_to_stream
function iteratorToStream(iterator: any) {
  return new ReadableStream({
    async pull(controller) {
      const { value, done } = await iterator.next()

      if (done) {
        controller.close()
      } else {
        controller.enqueue(value)
      }
    },
  })
}

function sleep(time: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, time)
  })
}

const encoder = new TextEncoder()

enum DeploymentStatus {
  NOT_STARTED = 'NOT_STARTED',
  IN_PROGRESS = 'IN_PROGRESS',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
}

type DeploymentDetails = {
  message: string
  status: DeploymentStatus
}

const mockGetDeploymentStatus = async (deploymentTag: string): Promise<DeploymentDetails> => {
  return Promise.resolve({
    message: `Deployment in progress for ${deploymentTag}`,
    status: DeploymentStatus.IN_PROGRESS,
  });
};

async function* makeIterator() {
  yield encoder.encode('Compiling Dockerfile...')
  await exec('nixpacks plan ./temp > ./temp/build-plan.json');
  await exec('nixpacks build --config build-plan.json ./temp');
  yield encoder.encode('Building OCI image...')
  await sleep(5000)
  yield encoder.encode('Deployed...')
}

export async function GET(
  request: NextRequest,
  { params: { tag } }: { params: { tag: string } },
) {
  const authStatus = await clerkClient.authenticateRequest({ request });
  if (!authStatus.isSignedIn) {
    return new Response(
      JSON.stringify({ message: "Unauthorized" }),
      { status: 401 },
    );
  }

  const { status } = await mockGetDeploymentStatus(tag);

  if (status === DeploymentStatus.SUCCESS || status === DeploymentStatus.FAILED) {
    return new Response(
      JSON.stringify({
        message: status,
      }),
      { status: 200 },
    )
  }

  const iterator = makeIterator()
  const stream = iteratorToStream(iterator)

  return new Response(stream)
}
