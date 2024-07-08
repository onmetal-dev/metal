import { db } from "@/app/server/db";
import {
  Application,
  ApplicationConfig,
  ApplicationConfigInsert,
  ApplicationConfigVersionData,
  Build,
  BuildArtifact,
  BuildInsert,
  Builder,
  Databases,
  Dependencies,
  Deployment,
  DeploymentInsert,
  Environment,
  External,
  HealthCheck,
  HetznerCluster,
  Port,
  Ports,
  Resources,
  Source,
  applicationConfigs,
  applicationVersion,
  builds,
  deployments,
} from "@/app/server/db/schema";
import * as deploy from "@/lib/deploy";
import { parseSsOutput } from "@/lib/ss";
import { tracedExec } from "@/lib/tracedExec";
import { getUser, idSchema, responseSpecs, userTeams } from "@api/shared";
import { createRoute, z, type OpenAPIHono } from "@hono/zod-openapi";
import { App } from "cdk8s";
import chalk from "chalk";
import { eq } from "drizzle-orm";
import { existsSync, symlinkSync } from "fs";
import { type Context } from "hono";
import { streamText } from "hono/streaming";
import { StreamingApi } from "hono/utils/stream";
import { spawn } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "path";
import { rimraf } from "rimraf";
import tmp from "tmp";
import * as uuidBase62 from "uuid-base62";

const bodyUp = z
  .object({
    teamId: idSchema.openapi({
      example: "3OHY5rQEfrc1vOpFrJ9q3r",
    }),
    envId: idSchema.openapi({
      example: "3OHY5rQEfrc1vOpFrJ9q3r",
    }),
    appId: idSchema.openapi({
      example: "3OHY5rQEfrc1vOpFrJ9q3r",
    }),
    archive: z.instanceof(File).openapi({ format: "binary", type: "string" }),
  })
  .openapi({ required: ["archive", "teamId", "envId", "appId"] });

type BodyUp = z.infer<typeof bodyUp>;

export default function upRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "post",
      operationId: "up",
      path: "/up",
      request: {
        body: {
          content: {
            "multipart/form-data": {
              schema: bodyUp,
            },
          },
        },
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: {
          description: "Deployment created",
          content: {
            "text/plain": {
              schema: {
                type: "string",
                example: "...deployment feedback...",
              },
            },
          },
        },
        400: responseSpecs[400],
        401: responseSpecs[401],
        404: responseSpecs[404],
      },
    }),
    // @ts-ignore
    async (c: Context) => {
      const user = getUser(c);
      const body = (await c.req.parseBody()) as BodyUp;
      const archive = body.archive;
      const teamId = body.teamId;
      const envId = body.envId;
      const appId = body.appId;

      const teams = await userTeams(user.id);
      const team = teams.find((t) => t.id === teamId);
      if (!team) {
        return c.json(
          { error: "unknown_team", message: `Unknown team ID ${teamId}` },
          400
        );
      }
      const env: Environment | undefined =
        await db.query.environments.findFirst({
          where: (t, { eq }) => eq(t.id, envId),
        });
      if (!env) {
        return c.json(
          {
            error: "unknown_environment",
            message: `Unknown environment ID ${envId}`,
          },
          400
        );
      }
      const app: Application | undefined =
        await db.query.applications.findFirst({
          where: (t, { eq }) => eq(t.id, appId),
        });
      if (!app) {
        return c.json(
          { error: "unknown_app", message: `Unknown app ID ${appId}` },
          400
        );
      }

      const clusters: HetznerCluster[] =
        await db.query.hetznerClusters.findMany({
          where: (c, { eq }) => eq(c.teamId, teamId),
        });
      if (clusters.length === 0) {
        return c.json(
          {
            error: "no_clusters",
            message: `No clusters found for team ${teamId}`,
          },
          400
        );
      }

      const tmpFileForArchive = tmp.fileSync();
      const tmpDirForExtraction = tmp.dirSync();
      const tmpDirForDockerConfig = tmp.dirSync();
      const arrayBuffer = await archive.arrayBuffer();
      writeFileSync(tmpFileForArchive.name, Buffer.from(arrayBuffer));
      return streamText(c, async (stream) => {
        stream.writeln(chalk.green("Extracting archive..."));
        await extractArchive({
          archivePath: tmpFileForArchive.name,
          destination: tmpDirForExtraction.name,
          stream,
        });

        // find/create application config
        // in the future, take in a param at this endpoint that is the rel path to app config file
        // and use that to find/create application version (will make it so you can customize stuff and have > 1 app in a repo)
        stream.writeln(chalk.green("Determining application config..."));
        const appConfig: ApplicationConfig = await findCreateApplicationConfig({
          appId: app.id,
          teamId: teamId,
        });

        const buildId = uuidBase62.v4();
        const buildInsert: BuildInsert = {
          id: buildId,
          teamId: teamId,
          applicationId: appId,
          applicationConfigId: appConfig.id,
          status: "running",
          logs: "",
          artifacts: [],
        };
        await db
          .insert(builds)
          // @ts-ignore
          .values(buildInsert)
          .returning();

        stream.writeln(chalk.green("Generating Dockerfile with nixpacks..."));
        await nixpacksBuild({
          appConfig,
          appPath: tmpDirForExtraction.name,
          imageName: app.name,
          destination: tmpDirForExtraction.name,
          stream,
        });

        stream.writeln(chalk.green("Configuring docker..."));
        const clusters = await db.query.hetznerClusters.findMany({
          where: (c, { eq }) => eq(c.teamId, teamId),
        });
        await constructDockerConfigDirForBuild({
          dir: tmpDirForDockerConfig.name,
          clusters,
        });

        stream.writeln(chalk.green("Building docker image..."));
        const tags: string[] = clusters.map(
          (c) => `registry.${c.name}.up.onmetal.dev/${app.name}:${buildId}`
        );
        await dockerBuildx({
          configDir: `${tmpDirForDockerConfig.name}/.docker`,
          builderName: process.env.CLOUD_BUILDER_NAME!,
          workingDir: tmpDirForExtraction.name,
          dockerfilePath: "./.nixpacks/Dockerfile",
          tags,
          stream,
        });

        const build: Build = await db
          .update(builds)
          .set({
            status: "completed",
            logs: "todo: save these",
            artifacts: clusters.map((c) => ({
              image: {
                repository: `registry.${c.name}.up.onmetal.dev`,
                name: app.name,
                tag: buildId,
              },
            })),
          })
          .where(eq(builds.id, buildId))
          .returning()
          .then((data: Build[]) => {
            if (data.length !== 1 || !data[0]) {
              throw new Error("More than one build returned");
            }
            return data[0];
          });
        const deploymentId = uuidBase62.v4();
        const deploymentInsert: DeploymentInsert = {
          id: deploymentId,
          teamId,
          creatorId: user.id,
          applicationId: appId,
          applicationConfigId: appConfig.id,
          environmentId: envId,
          buildId: buildId,
          variables: [],
          type: "deploy",
          rolloutStatus: "deploying",
        };
        await db
          .insert(deployments)
          // @ts-ignore
          .values(deploymentInsert);
        // for whatever reason the .returning() version of insert returns a deployment object with an undefined deployment.id, so run a query
        const deployment: Deployment | undefined =
          await db.query.deployments.findFirst({
            where: (d, { eq }) => eq(d.id, deploymentId),
          });
        if (!deployment) {
          throw new Error("Deployment not found");
        }
        await performDeployment({
          app,
          appConfig,
          build,
          deployment,
          clusters,
          stream,
        });

        for (const tmpDir of [
          tmpFileForArchive,
          tmpDirForExtraction,
          tmpDirForDockerConfig,
        ]) {
          rimraf.sync(tmpDir.name);
        }
      });
    }
  );
}

interface PerformDeploymentOptions {
  app: Application;
  appConfig: ApplicationConfig;
  build: Build;
  deployment: Deployment;
  clusters: HetznerCluster[];
  stream: StreamingApi;
}

async function performDeployment({
  app,
  appConfig,
  build,
  deployment,
  clusters,
  stream,
}: PerformDeploymentOptions) {
  if (clusters.length === 0) {
    throw new Error("performDeployment must have a list of clusters");
  }

  // all external ports must have a corresponding referent in appConfig.ports
  for (const ep of appConfig.external) {
    if (!appConfig.ports.find((p: Port) => p.name === ep.portName)) {
      throw new Error(
        `External port ${ep.name} not found in application config ports`
      );
    }
  }

  // all clusters must have a build artifact with image.repository === cluster.registry
  const buildArtifactForCluster = new Map<string, BuildArtifact>();
  for (const cluster of clusters) {
    const artifact: BuildArtifact | undefined = build.artifacts.find(
      (a: BuildArtifact) =>
        a.image?.repository === `registry.${cluster.name}.up.onmetal.dev`
    );
    if (!artifact || !artifact.image) {
      throw new Error(
        `No build artifact for this cluster: ${
          cluster.name
        } ${build.artifacts.map((a: BuildArtifact) => a.image?.repository)}`
      );
    }
    buildArtifactForCluster.set(cluster.id, artifact);
  }

  for (const cluster of clusters) {
    const tmpDirForRollout = tmp.dirSync();
    writeFileSync(
      `${tmpDirForRollout.name}/kubeconfig`,
      cluster.kubeconfig!,
      "utf8"
    );

    // imagePullSecret must be in the same namespace as the pod, so make a copy
    await copySecret({
      from: {
        kubeconfig: `${tmpDirForRollout.name}/kubeconfig`,
        namespace: "default",
        secretName: `regcred-${cluster.name}`,
      },
      to: {
        kubeconfig: `${tmpDirForRollout.name}/kubeconfig`,
        namespace: app.name,
        secretName: `regcred-${cluster.name}`,
      },
    });

    const cdkApp = new App();
    const namespace = new deploy.Namespace(cdkApp, `${app.name}-namespace`, {
      namespace: app.name,
    });
    const argoDeployment = new deploy.ArgoRollout(cdkApp, app.name, {
      app,
      appConfig,
      labels: { app: app.name, deploymentId: deployment.id },
      buildArtifacts: build.artifacts,
      clusterName: cluster.name,
    });
    argoDeployment.addDependency(namespace);
    writeFileSync(`${tmpDirForRollout.name}/deploy.yaml`, cdkApp.synthYaml());

    await tracedExec({
      spanName: "rollout-apply",
      command: `KUBECTL_APPLYSET=true kubectl --kubeconfig ${tmpDirForRollout.name}/kubeconfig apply -f ${tmpDirForRollout.name}/deploy.yaml --prune --applyset=configmaps/${app.name} -n ${app.name}`,
    });

    stream.writeln(
      chalk.green(
        `deployment in progress! Once app is up, visit https://${app.name}.${cluster.name}.up.onmetal.dev`
      )
    );

    db.update(deployments)
      .set({ rolloutStatus: "deploying" })
      .where(eq(deployments.id, deployment.id));
    //rimraf.sync(tmpDirForRollout.name);
  }

  // in the first cluster, perform a port check
  try {
    await portCheck({
      cluster: clusters[0]!,
      namespace: app.name,
      deployment,
      appConfig,
    });
  } catch (e) {
    if (e instanceof PortCheckError) {
      stream.writeln(chalk.red(`Port check failed: ${e.message}`));
      stream.writeln(
        chalk.red(`Expected ports: ${e.details.expectedPorts.join(", ")}`)
      );
      stream.writeln(
        chalk.red(`Detected ports: ${e.details.detectedPorts.join(", ")}`)
      );
      stream.writeln(
        "Possible reason: if the app does not have a config file, then we assume that the app respects a PORT environment variable that we inject, which is by default port 80"
      );
    } else {
      throw e; // rethrow if it's not a PortCheckError
    }
  }
}

interface PortCheckOptions {
  cluster: HetznerCluster;
  deployment: Deployment;
  namespace: string;
  appConfig: ApplicationConfig;
}

class PortCheckError extends Error {
  constructor(
    message: string,
    public details: { expectedPorts: number[]; detectedPorts: number[] }
  ) {
    super(message);
    this.name = "PortCheckError";
  }
}

// portCheck makes sure that the app is up and listening on the correct ports
async function portCheck({
  cluster,
  deployment,
  namespace,
  appConfig,
}: PortCheckOptions): Promise<void> {
  if (!cluster.kubeconfig) {
    throw new Error(`Cluster ${cluster.name} has no kubeconfig`);
  }
  const tmpFileForPortCheck = tmp.fileSync();
  writeFileSync(`${tmpFileForPortCheck.name}`, cluster.kubeconfig!, "utf8");
  let { stdout: podName } = await tracedExec({
    spanName: "get-pod-name",
    command: `kubectl --kubeconfig ${tmpFileForPortCheck.name} get pods -n ${namespace} -l deploymentId=${deployment.id} -o name | head -n1 | sed 's/pod\\///'`,
  });
  podName = podName.trim();
  if (podName === "") {
    throw new Error(
      `No pod found in namespace ${namespace} with label deploymentId=${deployment.id}`
    );
  }
  await tracedExec({
    spanName: "wait-for-pod-to-be-running",
    command: `kubectl --kubeconfig ${tmpFileForPortCheck.name} wait --for=condition=ready pod -n ${namespace} ${podName}`,
  });
  const { stdout: ssOutput } = await tracedExec({
    spanName: "get-ss-output",
    command: `kubectl --kubeconfig ${tmpFileForPortCheck.name} exec -n ${namespace} ${podName} -- ss -tulpn`,
  });
  tmpFileForPortCheck.removeCallback();
  const expectedPorts: number[] = appConfig.ports.map((p) => p.port);
  const detectedPorts: number[] = parseSsOutput(ssOutput).map(
    (c) => c.localPort
  );
  const detectedPortsSet: Set<number> = new Set(detectedPorts);
  const missingPortsSet: Set<number> = new Set(
    expectedPorts.filter((port) => !detectedPortsSet.has(port))
  );
  if (missingPortsSet.size > 0) {
    throw new PortCheckError("Expected ports do not match detected ports", {
      expectedPorts,
      detectedPorts,
    });
  }
}

async function copySecret({
  from,
  to,
}: {
  from: {
    kubeconfig: string;
    namespace: string;
    secretName: string;
  };
  to: {
    kubeconfig: string;
    namespace: string;
    secretName: string;
  };
}): Promise<void> {
  const { stdout: secretYaml } = await tracedExec({
    spanName: "get-secret-yaml",
    command: `kubectl --kubeconfig ${from.kubeconfig} get secret -n ${from.namespace} ${from.secretName} -o yaml`,
  });

  const updatedSecretYaml = secretYaml
    .replace(/namespace: .*/, `namespace: ${to.namespace}`)
    .replace(/^\s+uid: .*?\n/m, "")
    .replace(/^\s+resourceVersion: .*?\n/m, "")
    .replace(/^\s+creationTimestamp: .*?\n/m, "");

  const tmpDir = tmp.dirSync();
  const tmpFile = path.join(tmpDir.name, "secret.yaml");
  writeFileSync(tmpFile, updatedSecretYaml);

  await tracedExec({
    spanName: "apply-secret",
    spanAttributes: {},
    command: `kubectl --kubeconfig ${to.kubeconfig} apply -f ${tmpFile}`,
  });
}

interface FindCreateApplicationConfigOptions {
  teamId: string;
  appId: string;
}

async function findCreateApplicationConfig({
  teamId,
  appId,
}: FindCreateApplicationConfigOptions): Promise<ApplicationConfig> {
  const appConfigVersionData: ApplicationConfigVersionData = {
    teamId,
    applicationId: appId,
    source: {
      type: "upload" /* upload: { path: "...", hash: "..." } <- todo: store the location of the upload in object storage, but exclude this from app config version computation. or do we care? */,
    } as Source,
    builder: {
      type: "nixpacks" /* nixpacks: {} <- todo: if user specifies these in custom config */,
      nixpacks: {
        phases: {
          setup: {
            nixPkgs: ["...", "iproute2"],
            nixLibs: ["..."],
            nixOverlays: ["..."],
            aptPkgs: ["..."],
            dependsOn: ["..."],
            cacheDirectories: ["..."],
            onlyIncludeFiles: ["..."],
            paths: ["..."],
          },
        },
      },
    } as Builder,
    ports: [{ name: "http", proto: "http", port: 80 }] as Ports,
    external: [
      { name: "https", portName: "http", proto: "https", port: 443 },
    ] as External,
    healthCheck: { proto: "http", path: "/", portName: "http" } as HealthCheck,
    dependencies: [] as Dependencies,
    databases: [] as Databases,
    resources: { cpu: 0.25, memory: "256M" } as Resources,
  };
  const appConfigVersion = applicationVersion(appConfigVersionData);
  const existingAppConfig = await db.query.applicationConfigs.findFirst({
    where: (c, { eq, and }) =>
      and(eq(c.applicationId, appId), eq(c.version, appConfigVersion)),
  });
  if (existingAppConfig) {
    return existingAppConfig;
  }
  const insert: ApplicationConfigInsert = {
    ...appConfigVersionData,
    version: appConfigVersion,
  };
  const inserted: ApplicationConfig[] = await db
    .insert(applicationConfigs)
    // @ts-ignore... it can't figure this out
    .values(insert)
    .returning();
  return inserted[0]!;
}

interface DockerBuildxOptions {
  configDir: string;
  builderName: string;
  dockerfilePath: string;
  tags: string[];
  workingDir: string;
  stream: StreamingApi;
}

async function dockerBuildx({
  configDir,
  builderName,
  workingDir,
  dockerfilePath,
  tags,
  stream,
}: DockerBuildxOptions): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn(
      "docker",
      [
        "--config",
        configDir,
        "buildx",
        "build",
        "--builder",
        builderName,
        ".",
        "-f",
        dockerfilePath,
        ...tags.map((t) => ["-t", t]).flat(),
        "--push",
      ],
      {
        cwd: workingDir,
      }
    );
    proc.stdout.on("data", relayData.bind(null, stream));
    proc.stderr.on("data", relayData.bind(null, stream));
    proc.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`docker buildx failed with code ${code}`));
        return;
      }
      resolve();
    });
  });
}

interface ClusterDockerRegistryAuth {
  authsKey: string;
  authsValue: any;
}

async function clusterDockerRegistryAuth(
  cluster: HetznerCluster
): Promise<ClusterDockerRegistryAuth> {
  if (!cluster.kubeconfig) {
    throw new Error(`Cluster ${cluster.name} has no kubeconfig`);
  }
  const kubeconfigFile = tmp.fileSync();
  writeFileSync(kubeconfigFile.name, cluster.kubeconfig, "utf8");
  const { stdout: clusterDockerConfig } = await tracedExec({
    spanAttributes: {
      clusterName: cluster.name,
    },
    spanName: "cluster-get-regcred-secret",
    command: `kubectl --kubeconfig ${kubeconfigFile.name} get secret regcred-${cluster.name} -o json | jq -r '.data[".dockerconfigjson"]' | base64 --decode`,
  });
  const authsKey = `registry.${cluster.name}.up.onmetal.dev`;
  const authsValue = JSON.parse(clusterDockerConfig).auths[authsKey];
  if (!authsValue) {
    throw new Error(`No value found for ${authsKey} in ${clusterDockerConfig}`);
  }
  return { authsKey, authsValue };
}

interface ConstructDockerConfigDirForBuild {
  dir: string;
  clusters: HetznerCluster[];
}

async function constructDockerConfigDirForBuild({
  dir,
  clusters,
}: ConstructDockerConfigDirForBuild) {
  const dockerConfig: {
    auths: Record<string, any>;
  } = {
    auths: {},
  };
  const clusterAuths = await Promise.all(
    clusters.map(clusterDockerRegistryAuth)
  );
  for (const clusterAuth of clusterAuths) {
    dockerConfig.auths[clusterAuth.authsKey] = clusterAuth.authsValue;
  }
  // docker cloud builder auth
  dockerConfig.auths["https://index.docker.io/v1/"] = {
    auth: Buffer.from(
      `${process.env.CLOUD_BUILDER_USERNAME}:${process.env.CLOUD_BUILDER_PASSWORD}`
    ).toString("base64"),
  };
  mkdirSync(`${dir}/.docker`);
  writeFileSync(
    `${dir}/.docker/config.json`,
    JSON.stringify(dockerConfig, null, 2),
    "utf8"
  );

  // "install" docker buildx into the config dir by iterating through some potential locations for a system-wide buildx install
  let success = false;
  for (const dirWithBuildx of [
    "/usr/local/lib/docker/cli-plugins",
    "/usr/lib/docker/cli-plugins",
    "/Applications/Docker.app/Contents/Resources/cli-plugins",
  ]) {
    const buildxPath = `${dirWithBuildx}/docker-buildx`;
    if (existsSync(buildxPath)) {
      mkdirSync(`${dir}/.docker/cli-plugins`);
      symlinkSync(buildxPath, `${dir}/.docker/cli-plugins/docker-buildx`);
      success = true;
      break;
    }
  }
  if (!success) {
    throw new Error("Could not find docker-buildx");
  }
  await tracedExec({
    command: `docker --config ${dir}/.docker buildx create --driver cloud onmetal/arm-builder --name ${process.env.CLOUD_BUILDER_NAME}`,
    spanName: "docker-buildx-create",
  });
  return;
}

async function relayData(stream: StreamingApi, data: Buffer) {
  stream.write(data.toString());
}

interface NixpacksBuildOptions {
  appConfig: ApplicationConfig;
  appPath: string;
  imageName: string;
  destination: string;
  stream: StreamingApi;
}
async function nixpacksBuild({
  appConfig,
  appPath,
  imageName,
  destination,
  stream,
}: NixpacksBuildOptions) {
  const nixpacksConfig = tmp.fileSync({ postfix: ".json" });
  writeFileSync(
    nixpacksConfig.name,
    JSON.stringify(appConfig.builder.nixpacks, null, 2),
    "utf8"
  );
  return new Promise<void>((resolve, reject) => {
    const proc = spawn("nixpacks", [
      "build",
      appPath,
      "--config",
      nixpacksConfig.name,
      "--name",
      imageName,
      "--out",
      destination,
    ]);
    proc.stdout.on("data", relayData.bind(null, stream));
    proc.stderr.on("data", relayData.bind(null, stream));
    proc.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`nixpacks build failed with code ${code}`));
        return;
      }
      resolve();
    });
  });
}

interface ExtractArchiveOptions {
  archivePath: string;
  destination: string;
  stream: StreamingApi;
}
async function extractArchive({
  archivePath,
  destination,
  stream,
}: ExtractArchiveOptions): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    const proc = spawn("tar", ["xzfv", archivePath, "-C", destination]);
    proc.stdout.on("data", relayData.bind(null, stream));
    proc.stderr.on("data", relayData.bind(null, stream));
    proc.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`Archive extraction failed with code ${code}`));
        return;
      }
      resolve();
    });
    proc.on("error", (error) => {
      stream.writeln(chalk.red("Error extracting archive."));
      stream.writeln(error.message);
      reject(error);
    });
  });
}
