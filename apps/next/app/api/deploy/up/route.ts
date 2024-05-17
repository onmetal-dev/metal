import { getDirectoryNameForUps } from "@/app/server/util/functions";
import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { exec as execCallbackBased, spawn } from "node:child_process";
import { createWriteStream, existsSync, mkdirSync, writeFileSync } from "node:fs";
import { Writable } from "node:stream";
import { dirSync } from "tmp";
import { promisify } from "node:util";
import { NixpackPlan } from "@/types/deployment";

export const dynamic = "force-dynamic";

const exec = promisify(execCallbackBased);
const upsDirectory = getDirectoryNameForUps();
const org = process.env.DOCKERHUB_ORG;
const repo = process.env.DOCKERHUB_REPO;
const cloudBuilderName = process.env.CLOUD_BUILDER_NAME;

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

const encoder = new TextEncoder();

async function* makeIterator(buildTag: string) {
  yield encoder.encode(`--> Deployment started. Tag is ${buildTag}.`);

  yield encoder.encode('[<FE>] Planning build...');
  const tempDirName = `./${upsDirectory}/${buildTag}`;
  const { stdout: stdoutJson } = await exec(`nixpacks plan ${tempDirName}`);
  const plan = JSON.parse(stdoutJson) as NixpackPlan;
  // Filtering out 'npm-9_x' because for some reason it's not listed on:
  // https://search.nixos.org/packages
  plan.phases.setup.nixPkgs =
    plan.phases.setup.nixPkgs.filter(nixPackage => nixPackage !== 'npm-9_x');
  writeFileSync(`${tempDirName}/build-plan.json`, JSON.stringify(plan, null, 2), "utf8");
  await sleep(1000);

  yield encoder.encode('[<FE>] Generating Dockerfile...');
  await exec(`nixpacks build ${tempDirName} --name ${buildTag} --config build-plan.json --out ${tempDirName}`);
  await exec(`mv ${tempDirName}/.nixpacks/Dockerfile ${tempDirName}/Dockerfile`);

  yield encoder.encode('[<FE>] Building OCI image...');
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
      // Resorted to logging for now as an initial attempt to use "yield" here didn't work, apparently because it was an inner generator (i.e. inner function*). Will try again.
      // yield encoder.encode(`--> [DOCKER BUILD] ${data}`);
    });
  });

  yield encoder.encode('[<FE>] OCI image built...');
}

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
    let isFirstDataChunk = true;
    extractionStream.stderr.on('data', (data: Buffer) => {
      if (isFirstDataChunk) {
        isFirstDataChunk = false;
        console.log('***** METAL: Extract files *****');
      }

      console.log(`${data}`);
    });
    extractionStream.on('exit', (code) => {
      if (code !== 0) {
        reject(new Error(`Tarball extraction failed with code ${code}`));
        return;
      }

      console.log("Tarball extracted");
      resolve();
    });
    extractionStream.on('error', (error) => {
      reject(error);
    })
  });

  const iterator = makeIterator(tag)
  const stream = iteratorToStream(iterator)

  return new Response(stream)
}
