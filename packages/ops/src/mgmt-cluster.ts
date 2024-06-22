import { confirm } from "@inquirer/prompts";
import * as k8s from "@kubernetes/client-node";
import type * as hcloud from "@metal/hcloud";
import { $ } from "bun";
import chalk from "chalk";
import { Command } from "commander";
import { existsSync, mkdirSync, readdirSync, writeFileSync } from "fs";
import createClient from "openapi-fetch";
import { table } from "table";
import * as tmp from "tmp";
import { z } from "zod";
import { fromError } from "zod-validation-error";

function saveKubeconfig({ name, content }: { name: string; content: string }) {
  mkdirSync("./.kubeconfigs", { recursive: true });
  writeFileSync(`./.kubeconfigs/${name}.yaml`, content, {
    mode: 0o600, // avoid warnings about file being group and world readable
  });
}
function saveClusterYaml({ name, content }: { name: string; content: string }) {
  mkdirSync("./.clusteryamls", { recursive: true });
  writeFileSync(`./.clusteryamls/${name}.yaml`, content);
}

async function ensureCommands(cmds: string[]) {
  let errors: (string | null)[] = await Promise.all(
    cmds.map(async (cmd) => {
      const { exitCode } = await $`which ${cmd}`.quiet().nothrow();
      if (exitCode !== 0) {
        return cmd;
      }
      return null;
    })
  );
  errors = errors.filter((e) => e !== null);
  if (errors.length > 0) {
    console.error(
      chalk.red(`Required command line tools not found: ${errors.join(", ")}`)
    );
    process.exit(1);
  }
}

async function localKindCluster(
  name: string
): Promise<{ kubeconfigOpt: string; kubeconfig: string }> {
  console.log(chalk.green(`Creating local kind cluster ${name}`));
  await $`kind create cluster -n ${name}`;
  const kubeconfig = tmp.fileSync();
  await $`kind get kubeconfig -n ${name} > ${kubeconfig.name}`.quiet();
  console.log(chalk.green(`Initializing cluster-api in ${name}`));
  const kubeconfigOpt = `--kubeconfig=${kubeconfig.name}`;
  await $`clusterctl init ${kubeconfigOpt} --core cluster-api --bootstrap kubeadm --control-plane kubeadm --infrastructure hetzner`;
  return { kubeconfigOpt, kubeconfig: kubeconfig.name };
}

export default function mgmtCluster(program: Command) {
  program.description("Manage management clusters");

  const createHcloudSchema = z.object({
    location: z
      .enum(["fsn1", "nbg1", "hel1"])
      .describe("Location: fsn1, nbg1, or hel1"),
    controlPlaneMachineType: z
      .enum(["cax11", "cax21", "cax31", "cax41"])
      .describe("Control plane machine type: cax11, cax21, cax31, or cax41"),
    workerMachineType: z
      .enum(["cax11", "cax21", "cax31", "cax41"])
      .describe("Worker machine type: cax11, cax21, cax31, or cax41"),
    hcloudToken: z.string().describe("Hetzner Cloud API token"),
    sshKey: z.string().describe("SSH key name"),
    robotUser: z.string().describe("Hetzner Robot user"),
    robotPassword: z.string().describe("Hetzner Robot password"),
    kubernetesVersion: z
      .string()
      .describe("Kubernetes version")
      .regex(
        /^\d+\.\d+\.\d+$/,
        "Invalid Kubernetes version. Must be of the form 1.28.4"
      ),
    controlPlaneMachineCount: z
      .string()
      .describe("Control plane machine count")
      .regex(/^\d+$/, "Invalid control plane machine count"),
    workerMachineCount: z
      .string()
      .describe("Worker machine count")
      .regex(/^\d+$/, "Invalid worker machine count"),
    sshPubPath: z
      .string()
      .describe("Path to SSH public key file")
      .refine(existsSync, "SSH public key file does not exist"),
    sshPrivPath: z
      .string()
      .describe("Path to SSH private key file")
      .refine(existsSync, "SSH private key file does not exist"),
    name: z
      .string()
      .describe("Cluster name")
      .regex(/^[a-z0-9-]+$/, "Invalid cluster name"),
  });
  type CreateHcloudOptions = z.infer<typeof createHcloudSchema>;
  function optionDescription(name: keyof CreateHcloudOptions): string {
    return createHcloudSchema.shape[name].description ?? "";
  }

  program
    .command("create")
    .argument("<name>", "Name for the cluster")
    .description(
      "Create a management cluster in Hetzner Cloud. Assumes you have a Hetzner account set up (project + api key + ssh key set up"
    )
    .requiredOption(
      "-l, --location [location]",
      optionDescription("location"),
      "fsn1"
    )
    .requiredOption(
      "--control-plane-machine-type [controlPlaneMachineType]",
      optionDescription("controlPlaneMachineType"),
      "cax11"
    )
    .requiredOption(
      "--worker-machine-type [workerMachineType]",
      optionDescription("workerMachineType"),
      "cax11"
    )
    .requiredOption(
      "-t, --hcloud-token [hcloudToken]",
      optionDescription("hcloudToken")
    )
    .requiredOption(
      "-s, --ssh-key [sshKey]",
      optionDescription("sshKey"),
      "default-ssh-key"
    )
    .requiredOption("--robot-user [robotUser]", optionDescription("robotUser"))
    .requiredOption(
      "--robot-password [robotPassword]",
      optionDescription("robotPassword")
    )
    .requiredOption(
      "-k, --kubernetes-version [kubernetesVersion]",
      optionDescription("kubernetesVersion"),
      "1.28.4"
    )
    .requiredOption(
      "--control-plane-machine-count [controlPlaneMachineCount]",
      optionDescription("controlPlaneMachineCount"),
      "1"
    )
    .requiredOption(
      "--worker-machine-count [workerMachineCount]",
      optionDescription("workerMachineCount"),
      "1"
    )
    .requiredOption(
      "--ssh-pub-path [sshPubPath]",
      optionDescription("sshPubPath")
    )
    .requiredOption(
      "--ssh-priv-path [sshPrivPath]",
      optionDescription("sshPrivPath")
    )
    .action(async (name: string, options: CreateHcloudOptions) => {
      options.name = name;
      const parsed = createHcloudSchema.safeParse(options);
      if (parsed.error) {
        console.error(
          chalk.red(
            `${fromError(parsed.error, {
              prefix: "Invalid options",
              prefixSeparator: ":\n- ",
              issueSeparator: "\n- ",
              includePath: false,
            }).toString()}`
          )
        );
        process.exit(1);
      }

      await ensureCommands(["kind", "kubectl", "clusterctl"]);

      const bootsrapClusterName = `${name}-bootstrap`;
      const { kubeconfigOpt } = await localKindCluster(bootsrapClusterName);
      const {
        location,
        controlPlaneMachineType,
        workerMachineType,
        hcloudToken,
        sshKey,
        robotUser,
        robotPassword,
        kubernetesVersion,
        controlPlaneMachineCount,
        workerMachineCount,
        sshPubPath,
        sshPrivPath,
      } = parsed.data;
      console.log(
        chalk.green(`Creating management cluster ${name} in Hetzner Cloud`)
      );
      $.env({
        ...process.env,
        HCLOUD_TOKEN: hcloudToken,
        HCLOUD_SSH_KEY: sshKey,
        HETZNER_ROBOT_USER: robotUser,
        HETZNER_ROBOT_PASSWORD: robotPassword,
        HETZNER_SSH_PUB_PATH: sshPubPath,
        HETZNER_SSH_PRIV_PATH: sshPrivPath,
        CLUSTER_NAME: name,
        HCLOUD_REGION: location,
        CONTROL_PLANE_MACHINE_COUNT: controlPlaneMachineCount,
        WORKER_MACHINE_COUNT: workerMachineCount,
        KUBERNETES_VERSION: kubernetesVersion,
        HCLOUD_CONTROL_PLANE_MACHINE_TYPE: controlPlaneMachineType,
        HCLOUD_WORKER_MACHINE_TYPE: workerMachineType,
      });

      await $`kubectl ${kubeconfigOpt} create secret generic hetzner --from-literal=hcloud=${hcloudToken} --from-literal=robot-user=${robotUser} --from-literal=robot-password=${robotPassword}`.quiet();
      await $`kubectl ${kubeconfigOpt} create secret generic robot-ssh --from-literal=sshkey-name=cluster --from-file=ssh-privatekey=${sshPrivPath} --from-file=ssh-publickey=${sshPubPath}`.quiet();
      await $`kubectl ${kubeconfigOpt} patch secret hetzner -p '{"metadata":{"labels":{"clusterctl.cluster.x-k8s.io/move":""}}}'`.quiet();
      await $`kubectl ${kubeconfigOpt} patch secret robot-ssh -p '{"metadata":{"labels":{"clusterctl.cluster.x-k8s.io/move":""}}}'`.quiet();

      console.log(
        chalk.green("Waiting for pods in bootstrap cluster to be ready")
      );
      for (const namespace of [
        "kube-system",
        "capi-system",
        "capi-kubeadm-bootstrap-system",
        "capi-kubeadm-control-plane-system",
        "caph-system",
        "cert-manager",
      ]) {
        await $`kubectl ${kubeconfigOpt} wait --for=condition=Ready pod --all --namespace ${namespace} --timeout=1m`;
      }

      console.log(chalk.green("Generating cluster yaml..."));
      const { stdout: clusterYaml } =
        await $`clusterctl generate cluster ${name} ${kubeconfigOpt} --flavor hcloud`.quiet();
      saveClusterYaml({ name, content: clusterYaml.toString() });
      console.log(chalk.green("Applying cluster yaml..."));
      await $`kubectl ${kubeconfigOpt} apply -f .clusteryamls/${name}.yaml`;
      while (true) {
        const { stdout } =
          await $`kubectl ${kubeconfigOpt} get cluster ${name} -ojson`.quiet();
        const json = JSON.parse(stdout.toString());
        const { controlPlaneReady, infrastructureReady } = json.status;
        console.log(
          chalk.green(
            `Cluster ${name} status: controlPlaneReady=${
              controlPlaneReady ?? "false"
            } infrastructureReady=${infrastructureReady ?? "false"}`
          )
        );
        if (controlPlaneReady && infrastructureReady) {
          break;
        }
        await $`sleep 5`;
      }

      // wait for control plane to come up by looping on kubectl get --raw='/readyz?verbose' until it returns exit code 0
      while (true) {
        const { exitCode } =
          await $`kubectl ${kubeconfigOpt} get --raw='/readyz?verbose'`
            .quiet()
            .nothrow();
        console.log(
          chalk.green(
            `Control plane ready: ${exitCode === 0 ? "true" : "false"}`
          )
        );
        if (exitCode === 0) {
          break;
        }
        await $`sleep 5`;
      }

      // pull down the new cluster kubeconfig so we can install some things
      await $`clusterctl get kubeconfig ${name} ${kubeconfigOpt} > .kubeconfigs/${name}.yaml`;
      const kubeconfigOptNew = `--kubeconfig=.kubeconfigs/${name}.yaml`;

      console.log(chalk.green("installing cni (cilium)"));
      const caphCiliumValues = tmp.fileSync();
      await $`curl -o ${caphCiliumValues.name} https://raw.githubusercontent.com/syself/cluster-api-provider-hetzner/main/templates/cilium/cilium.yaml`;
      await $`helm ${kubeconfigOptNew} repo add cilium https://helm.cilium.io/`;
      await $`helm ${kubeconfigOptNew} repo update cilium`;
      await $`helm ${kubeconfigOptNew} upgrade --install cilium cilium/cilium --version 1.14.4 --namespace kube-system -f ${caphCiliumValues.name}`;

      console.log(chalk.green("installing hcloud cloud controller manager"));
      // ccm requires a secret on the cluster named 'hcloud' that includes the network name in hetzner cloud
      // do it via a yaml file so that we specify base64 values and avoid mistakes escaping any special chars
      const secretYaml = tmp.fileSync();
      writeFileSync(
        secretYaml.name,
        `apiVersion: v1
kind: Secret
metadata:
  name: hcloud
  namespace: kube-system
type: Opaque
data: ${JSON.stringify({
          token: Buffer.from(hcloudToken).toString("base64"),
          "robot-user": Buffer.from(robotUser).toString("base64"),
          "robot-password": Buffer.from(robotPassword).toString("base64"),
          network: Buffer.from(name).toString("base64"), // this relies on caph's convention of naming network after the cluster
        })}
      `
      );
      await $`kubectl ${kubeconfigOptNew} apply -f ${secretYaml.name}`.quiet();
      await $`helm ${kubeconfigOptNew} repo add hcloud https://charts.hetzner.cloud`;
      await $`helm ${kubeconfigOptNew} repo update hcloud`;
      await $`helm ${kubeconfigOptNew} upgrade --install hccm hcloud/hcloud-cloud-controller-manager --version 1.19.0 --namespace kube-system`;

      console.log(chalk.green("installing hcloud csi"));
      await $`helm ${kubeconfigOptNew} upgrade --install hcloud-csi hcloud/hcloud-csi --version 2.6.0 --namespace kube-system`;

      // install cluster-api on the new cluster and move resources (aka the ones for this mgmt cluster into it)
      console.log(chalk.green("Initializing cluster-api on new cluster"));
      await $`clusterctl init ${kubeconfigOptNew}  --core cluster-api --bootstrap kubeadm --control-plane kubeadm --infrastructure hetzner`;
      console.log(
        chalk.green("moving resources from bootstrap cluster to new cluster")
      );
      await $`clusterctl move ${kubeconfigOpt} --to-kubeconfig .kubeconfigs/${name}.yaml`
        .text;

      // delete the bootsrap kind cluster
      console.log(
        chalk.green(`Deleting local kind cluster ${bootsrapClusterName}`)
      );
      await $`kind delete cluster -n ${bootsrapClusterName}`;

      console.log(
        chalk.green(
          `All finished! Try running \`kubectl --kubeconfig .kubeconfigs/${name}.yaml get all --all-namespaces\` to see the new things runninging in the new cluster`
        )
      );
    });

  // delete uses the hcloud api to nuke the servers. use with caution
  program
    .command("delete")
    .description("Delete a management cluster")
    .option("-y, --yes", "Confirm deletion")
    .argument("<name>", "Name of the cluster to delete")
    .action(async (name: string, { yes }) => {
      if (!readdirSync(".kubeconfigs").includes(`${name}.yaml`)) {
        console.log(chalk.red(`Cluster ${name} not found`));
        process.exit(1);
      }
      await ensureCommands(["kubectl"]);
      if (!yes) {
        const confirmed = await confirm({
          message: `Are you sure you want to delete cluster ${name}?`,
        });
        if (!confirmed) {
          console.log(chalk.red(`Deletion cancelled`));
          process.exit(0);
        }
      }

      // figure out external IPs for all nodes in the cluster. This will be how we find the servers to delete in hcloud
      const nodes =
        await $`kubectl --kubeconfig .kubeconfigs/${name}.yaml get nodes -ojson`.json();
      if (!nodes.items || nodes.items.length === 0) {
        console.log(chalk.red(`No nodes found for cluster ${name}`));
        process.exit(1);
      }
      const externalIps = nodes.items!.map(
        (node: any) =>
          node.status!.addresses!.find((a: any) => a.type === "ExternalIP")
            ?.address
      );

      const secret =
        await $`kubectl --kubeconfig .kubeconfigs/${name}.yaml get secret hcloud -n kube-system -ojson`.json();
      if (!secret.data?.token) {
        console.log(chalk.red(`No hcloud secret found for cluster ${name}`));
        process.exit(1);
      }
      const token = Buffer.from(secret.data!.token, "base64").toString();
      const client = createClient<hcloud.paths>({
        headers: { Authorization: `Bearer ${token}` },
        baseUrl: "https://api.hetzner.cloud/v1",
      });
      const { data, error } = await client.GET("/servers");
      if (error) {
        console.log(chalk.red(`Error getting servers: ${error}`));
        process.exit(1);
      }

      // figure out which servers map to the external IPs
      const serversToDelete = data!.servers!.filter((server) =>
        externalIps.includes(server.public_net!.ipv4!.ip)
      );
      if (serversToDelete.length !== externalIps.length) {
        console.log(
          chalk.red(
            `Found ${serversToDelete.length} servers to delete, but ${externalIps.length} nodes in the cluster`
          )
        );
        process.exit(1);
      }

      for (const server of serversToDelete) {
        console.log(chalk.green(`Deleting server ${server.name}`));
        const { error } = await client.DELETE("/servers/{id}", {
          params: { path: { id: server.id } },
        });
        if (error) {
          console.log(
            chalk.red(`Error deleting server ${server.name}: ${error}`)
          );
          process.exit(1);
        }
      }

      await $`rm -f .kubeconfigs/${name}.yaml`;
      console.log(chalk.green(`Deleted servers for cluster ${name}`));
    });

  program
    .command("list")
    .description("List clusters created")
    .action(async () => {
      let clusters: k8s.Cluster[] = [];
      const kubeconfigs = readdirSync(".kubeconfigs");
      for (const kubeconfig of kubeconfigs) {
        const kc = new k8s.KubeConfig();
        kc.loadFromFile(`.kubeconfigs/${kubeconfig}`);
        clusters.push(...kc.clusters);
      }
      console.log(
        table([
          ["Name", "Server"],
          ...clusters.map((cluster) => {
            return [cluster.name, cluster.server];
          }),
        ])
      );
    });
}
