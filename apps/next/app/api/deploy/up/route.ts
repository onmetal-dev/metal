import { DIR_FOR_CLIENT_DEPLOYMENTS } from "@/app/server/util/constants";
import { clerkClient } from "@clerk/nextjs";
import { type NextRequest } from "next/server";
import { exec as execCallbackBased, spawn } from "node:child_process";
import { createWriteStream, existsSync, mkdirSync, writeFileSync } from "node:fs";
import { Readable, Writable } from "node:stream";
import { dirSync } from "tmp";
import { promisify } from "node:util";
import { NixpackPlan } from "@/types/deployment";

export const dynamic = "force-dynamic";

const exec = promisify(execCallbackBased);
const org = process.env.DOCKERHUB_ORG;
const repo = process.env.DOCKERHUB_REPO;
const cloudBuilderName = process.env.CLOUD_BUILDER_NAME;

// This code below is heavily based on:
// https://nextjs.org/docs/app/building-your-application/routing/route-handlers#streaming

// https://developer.mozilla.org/docs/Web/API/ReadableStream#convert_async_iterator_to_stream
function iteratorToStream(iterator: AsyncGenerator<Uint8Array, void, unknown>) {
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

// Thanks ChatGPT!
async function* streamIterator(stderr: Readable) {
  for await (const chunk of stderr) {
    yield chunk.toString();
  }
}

async function* makeIterator(buildTag: string, fileName: string) {
  const { name: tempDirName } = dirSync({
    tmpdir: DIR_FOR_CLIENT_DEPLOYMENTS,
    name: buildTag,
  });
  const extractionStream = spawn("tar", ["xzfv", fileName, "-C", tempDirName]);
  const extractionPromise = new Promise<void>((resolve, reject) => {
    extractionStream.on("error", (error) => {
      console.error("Error: ", error);
      reject(error);
    });
    extractionStream.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`Tarball extraction failed with code ${code}`));
        return;
      }

      resolve();
    });
  });

  let isFirstDataFromExtraction = true;
  for await (const data of streamIterator(extractionStream.stderr)) {
    if (isFirstDataFromExtraction) {
      isFirstDataFromExtraction = false;
      console.log(`***** METAL: Extract files for ${buildTag} *****`);
    }

    yield encoder.encode(`${data}`);
  }

  await extractionPromise;

  yield encoder.encode(`[<metal>]Deployment started. Tag is ${buildTag}.[</metal>]`);

  yield encoder.encode("[<metal>]Planning build...[</metal>]");
  const { stdout: stdoutJson } = await exec(`nixpacks plan ${tempDirName}`);
  const plan = JSON.parse(stdoutJson) as NixpackPlan;
  // Filtering out "npm-9_x" because for some reason it's not listed on:
  // https://search.nixos.org/packages
  plan.phases.setup.nixPkgs =
    plan.phases.setup.nixPkgs.filter(nixPackage => nixPackage !== "npm-9_x");
  writeFileSync(`${tempDirName}/build-plan.json`, JSON.stringify(plan, null, 2), "utf8");
  await sleep(1000);

  yield encoder.encode("[<metal>]Generating Dockerfile...[</metal>]");
  await exec(`nixpacks build ${tempDirName} --name ${buildTag} --config build-plan.json --out ${tempDirName}`);
  await exec(`mv ${tempDirName}/.nixpacks/Dockerfile ${tempDirName}/Dockerfile`);

  yield encoder.encode("[<metal>]Building OCI image...[</metal>]");
  /* NOTE: the build takes place in the cloud, so a copy of the build image is
  not available locally. Using the "--push" flag works because Docker BuildCloud
  can directly export the built image to DockerHub. If you want to use the
  image locally, you'll have to replace "--push" with something like
  "--output type=docker". After all, "--push" is just shorthand for
  "--output type=registry". For more details, see:
  https://docs.docker.com/reference/cli/docker/buildx/build/#output
  */
  const hasCustomDockerConfig = existsSync(`${tempDirName}/.docker/config.json`);
  const dockerFlags = [
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
    dockerFlags.unshift("--config", `${tempDirName}/.docker/config.json`);
  }

  const dockerBuildStream = spawn("docker", dockerFlags);

  const dockerPromise = new Promise<void>((resolve, reject) => {
    dockerBuildStream.on("error", (err) => {
      console.error("Error: ", err);
      reject(err);
    });

    dockerBuildStream.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`Docker build failed with code ${code}`));
        return;
      }

      resolve();
    });
  });

  let isFirstDataFromDockerBuild = true;
  for await (const data of streamIterator(dockerBuildStream.stderr)) {
    if (isFirstDataFromDockerBuild) {
      isFirstDataFromDockerBuild = false;
      console.log(`***** METAL: Docker build for ${buildTag} *****`);
    }

    yield encoder.encode(`${data}`);
  }

  await dockerPromise;

  yield encoder.encode("[<metal>]OCI image built...[</metal>]");
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
  if (!existsSync(DIR_FOR_CLIENT_DEPLOYMENTS)) {
    mkdirSync(DIR_FOR_CLIENT_DEPLOYMENTS);
  }

  const fileName = `${tag}.gz`;
  const uploadedTarball = Writable.toWeb(
    createWriteStream(fileName, "binary")
  ) as WritableStream<Uint8Array>;
  await request.body.pipeTo(uploadedTarball);

  const iterator = makeIterator(tag, fileName)
  const stream = iteratorToStream(iterator)

  return new Response(stream)
}
