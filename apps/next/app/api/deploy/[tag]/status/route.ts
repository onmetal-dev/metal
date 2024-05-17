import { getDirectoryNameForUps } from "@/app/server/util/functions";
import { NixpackPlan } from "@/types/deployment";
import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { exec as execCallbackBased, spawn } from "node:child_process";
import { writeFileSync, existsSync } from "node:fs";
import { promisify } from "node:util";

const exec = promisify(execCallbackBased);
const org = process.env.DOCKERHUB_ORG;
const repo = process.env.DOCKERHUB_REPO;
const cloudBuilderName = process.env.CLOUD_BUILDER_NAME;
const upsDirectory = getDirectoryNameForUps();

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

async function* makeIterator(buildTag: string) {
  yield encoder.encode('Planning build...');
  const tempDirName = `./${upsDirectory}/${buildTag}`;
  const { stdout: stdoutJson } = await exec(`nixpacks plan ${tempDirName}`);
  const plan = JSON.parse(stdoutJson) as NixpackPlan;
  // Filtering out 'npm-9_x' because for some reason it's not listed on:
  // https://search.nixos.org/packages
  plan.phases.setup.nixPkgs =
    plan.phases.setup.nixPkgs.filter(nixPackage => nixPackage !== 'npm-9_x');
  writeFileSync(`${tempDirName}/build-plan.json`, JSON.stringify(plan, null, 2), "utf8");
  await sleep(1000);

  yield encoder.encode('Generating Dockerfile...');
  await exec(`nixpacks build ${tempDirName} --name ${buildTag} --config build-plan.json --out ${tempDirName}`);
  await exec(`mv ${tempDirName}/.nixpacks/Dockerfile ${tempDirName}/Dockerfile`);
  await sleep(1000);

  yield encoder.encode('Building OCI image...');
  /* NOTE: the build takes place in the cloud, so a copy of the build image is
  not available locally. Using the "--push" flag works because Docker BuildCloud
  can directly export the built image to DockerHub. If you want to use the
  image locally, you'll have to replace "--push" with something like
  "--output type=docker". After all, "--push" is just shorthand for
  "--output type=registry". For more details, see:
  https://docs.docker.com/reference/cli/docker/buildx/build/#output
  */
  const hasCustomDockerConfig = existsSync(`${tempDirName}/.docker/config.json`);
  await new Promise((resolve, reject) => {
    // Trust me, don't have spaces within any flag.
    let dockerFlags = [
      "buildx",
      "build",
      tempDirName,
      "--builder",
      cloudBuilderName || "",
      "--tag",
      `${org}/${repo}:${buildTag}`,
      "--push",
    ];

    if (hasCustomDockerConfig) {
      dockerFlags = [
        // The --config flag is for the docker command itself. Keep it at the head of the list.
        '--config',
        `${tempDirName}/.docker/config.json`,
        ...dockerFlags,
      ];
    }

    const dockerBuildStream = spawn("docker", dockerFlags);
    dockerBuildStream.on('error', (err) => {
      console.log('err', err);
      reject(err);
    });

    dockerBuildStream.on('close', (code) => {
      resolve(code);
    });

    let isFirstDataChunk = true;
    dockerBuildStream.stderr.on('data', (data: Buffer) => {
      if (isFirstDataChunk) {
        isFirstDataChunk = false;
        console.log('***** METAL: Docker build *****');
      }

      console.log(`${data}`);
      //Resorted to logging for now as an initial attempt to use "yield" here didn't work, apparently because it was an inner generator (i.e. inner function*). Will try again.
      // yield encoder.encode(`--> [DOCKER BUILD] ${data}`);
    });
  });

  yield encoder.encode('OCI image built...');
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

  // TODO Folusho: store and check the deployment status.
  const { status } = await mockGetDeploymentStatus(tag);

  if (status === DeploymentStatus.SUCCESS || status === DeploymentStatus.FAILED) {
    return new Response(
      JSON.stringify({
        message: status,
      }),
      { status: 200 },
    )
  }

  const iterator = makeIterator(tag)
  const stream = iteratorToStream(iterator)

  return new Response(stream)
}
