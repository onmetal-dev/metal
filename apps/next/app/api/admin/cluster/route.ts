import * as k8s from "@kubernetes/client-node";
import { dump } from "js-yaml";
import util from "util";
import { exec as execCB, ExecException } from "child_process";
const exec = util.promisify(execCB);
import tmp from "tmp";
import fs from "fs";
import { SpanStatusCode, trace } from "@opentelemetry/api";
import { serviceName } from "@/lib/constants";

export const dynamic = "force-dynamic"; // defaults to auto

const mgmtCluster: k8s.Cluster = {
  name: process.env.HETZNER_MANAGEMENT_CLUSTER_NAME!,
  server: process.env.HETZNER_MANAGEMENT_CLUSTER_SERVER!,
  caData: process.env.HETZNER_MANAGEMENT_CLUSTER_CA_DATA!,
};

const mgmtUser: k8s.User = {
  name: "management-cluster-admin",
  certData: process.env.HETZNER_MANAGEMENT_CLUSTER_USER_CERT_DATA!,
  keyData: process.env.HETZNER_MANAGEMENT_CLUSTER_USER_KEY_DATA!,
};

const mgmtContext: k8s.Context = {
  name: `${mgmtUser.name}@${mgmtCluster.name}`,
  user: mgmtUser.name,
  cluster: mgmtCluster.name,
};

const mgmtKC = new k8s.KubeConfig();
mgmtKC.loadFromOptions({
  clusters: [mgmtCluster],
  users: [mgmtUser],
  contexts: [mgmtContext],
  currentContext: mgmtContext.name,
});

const mgmtClient = mgmtKC.makeApiClient(k8s.CoreV1Api);

const mgmtKCString = dump(JSON.parse(mgmtKC.exportConfig()));
const mgmtKCFile = tmp.fileSync();
fs.writeFileSync(mgmtKCFile.name, mgmtKCString);

const testUser = {
  id: "testuserid",
};

const testUserHCloudProject = {
  apiToken: "SRqdiZKhmfm5PEVBzgBGxCaRuE92PfNNMFFBRPcD592EQmgULPFJa6M8szFqsGhx",
  name: "metal-dev",
};

const testUserHRobot = {
  webserviceUser: "#ws+tRUWEpEz",
  webservicePassword: "wH.Kiaxg3Aj6VEsh",
};

const testUserHCloudCluster = {
  name: "test-cluster-8276", // todo: make this an randomly-generate heroku-like string, e.g. flaming-spire-8276
  machineTypeControlPlane: "cpx11",
  machineTypeWorker: "cpx11",
  region: "ash",
  sshKey: "metal-dev-ssh", // todo: user has provided this, so verify this exists. Failure mode in cluster-api is hard to catch
};

// tracedExec is a function that runs a command and traces it
// it takes in the span name, attributes to set, and the command to run
// it returns true on success, false if the command fails
// it adds a lot of attributes in order to make debugging easier
// it returns very little (just true/false) since a lot of this information is sensitive and should not be included in user-facing error messages
async function tracedExec({
  spanName,
  spanAttributes,
  command,
}: {
  spanName: string;
  spanAttributes: Record<string, string>;
  command: string;
}) {
  return await trace
    .getTracer(serviceName)
    .startActiveSpan(spanName, async (span) => {
      span.setAttributes(spanAttributes);
      span.setAttributes({ command });
      try {
        const { stdout, stderr } = await exec(command);
        span.setAttributes({ stdout, stderr });
        span.end();
        return true;
      } catch (e: any) {
        const error = e as ExecException;
        span.setAttributes({
          code: error.code,
          stdout: error.stdout,
          stderr: error.stderr,
          message: error.message,
        });
        span.setStatus({ code: SpanStatusCode.ERROR });
        span.end();
        return false;
      }
    });
}

export async function POST(request: Request) {
  // test kubectl works
  if (
    !(await tracedExec({
      spanName: "exec-kubectl-get-cluster",
      spanAttributes: {},
      command: `KUBECONFIG=${mgmtKCFile.name} kubectl get cluster --all-namespaces`,
    }))
  ) {
    return Response.json(
      { message: "failed to test connectivity to mgmt cluster" },
      { status: 500 }
    );
  }

  return Response.json({ workflowId: "foo-bar" }, { status: 200 });
}

export async function GET(request: Request) {
  const spanAttributes = {
    user: testUser.id,
    "hcloud-project": testUserHCloudProject.name,
    "hcloud-cluster-name": testUserHCloudCluster.name,
  };

  // add user's creds as secrets to the mgmt cluster.
  // cluster-api-provider-hetzner is configured with references to them
  // use --dry-run=client -o yaml piped to kubectl apply in order to make it idempotent
  // see https://stackoverflow.com/questions/45879498/how-can-i-update-a-secret-on-kubernetes-when-it-is-generated-from-a-file
  const userHetznerSecretName = `${testUser.id}-hetzner`;
  if (testUserHRobot.webservicePassword.includes("'")) {
    // this is a user-provided value so if we allow ' we could allow for RCE in the kubectl command
    throw new Error("webservicePassword must not contain a single quote");
  }
  if (
    !(await tracedExec({
      spanName: "exec-kubectl-create-secret",
      spanAttributes,
      command: `KUBECONFIG=${mgmtKCFile.name} kubectl create secret generic ${userHetznerSecretName} --save-config --dry-run=client -o yaml --from-literal=hcloud=${testUserHCloudProject.apiToken} --from-literal='robot-user=${testUserHRobot.webserviceUser}' --from-literal='robot-password=${testUserHRobot.webservicePassword}' | KUBECONFIG=${mgmtKCFile.name} kubectl apply -f -`,
    }))
  ) {
    return Response.json(
      { message: "failed to create secrets in management cluster" },
      { status: 500 }
    );
  }

  const clusterConfigFile = tmp.fileSync();
  if (
    !(await tracedExec({
      spanName: "exec-clusterctl-generate-cluster",
      spanAttributes,
      command: `HCLOUD_CONTROL_PLANE_MACHINE_TYPE=${testUserHCloudCluster.machineTypeControlPlane} HCLOUD_REGION=${testUserHCloudCluster.region} HCLOUD_SSH_KEY=${testUserHCloudCluster.sshKey} HCLOUD_WORKER_MACHINE_TYPE=${testUserHCloudCluster.machineTypeWorker} KUBECONFIG=${mgmtKCFile.name} clusterctl generate cluster ${testUserHCloudCluster.name} --flavor hcloud > ${clusterConfigFile.name}`,
    }))
  ) {
    return Response.json(
      { message: "failed to generate cluster config" },
      { status: 500 }
    );
  }
  let clusterConfig = fs.readFileSync(clusterConfigFile.name, "utf8");

  // Find / replace the hetzerSecretRef. By default it is `hetzner` but to support multiple users we need to change it to `<userid>-hetzner` (the naming convention used above in the kubectl create secret command).
  // We could do this via kustomize, but for now just replace "name: hetzner" in the string
  // throw an error if more than one occurence of name: hetzner
  const hetznerMatches = clusterConfig.match(/name: hetzner/g);
  if (hetznerMatches && hetznerMatches.length > 1) {
    throw new Error(
      `More than one occurence of name: hetzner... time to switch to kustomize`
    );
  }
  clusterConfig = clusterConfig.replace(
    /name: hetzner/,
    `name: ${userHetznerSecretName}`
  );
  fs.writeFileSync(clusterConfigFile.name, clusterConfig);

  if (
    !(await tracedExec({
      spanName: "exec-kubectl-apply-cluster-config",
      spanAttributes,
      command: `KUBECONFIG=${mgmtKCFile.name} kubectl apply -f ${clusterConfigFile.name}`,
    }))
  ) {
    return Response.json(
      { message: "failed to apply cluster config" },
      { status: 500 }
    );
  }

  console.log(`DEBUG: location of cluster config: ${clusterConfigFile.name}`);
  return Response.json({ message: "OK" });

  // wait for the control plane to be ready, and then grab its kubeconfig
  const timeout = 60 * 1000;
  const start = Date.now();
  let clusterKubeconfig: string | null = null;
  do {
    const { stdout } = await exec(
      `/bin/bash -c "KUBECONFIG=${mgmtKCFile.name} kubectl get kubeadmcontrolplane -ojson -l cluster.x-k8s.io/cluster-name=${testUserHCloudCluster.name}"`
    );
    const kubeadmControlPlanes = JSON.parse(stdout);
    if (
      kubeadmControlPlanes.items.length === 1 &&
      kubeadmControlPlanes.items[0].status?.ready === true
    ) {
      const { stdout } = await exec(
        `/bin/bash -c "KUBECONFIG=${mgmtKCFile.name} clusterctl get kubeconfig ${testUserHCloudCluster.name}"`
      );
      clusterKubeconfig = stdout;
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  } while (clusterKubeconfig === null && Date.now() - start < timeout);
  if (clusterKubeconfig === null) {
    throw new Error(`Timed out waiting for control plane to be ready`);
  }

  // todo: save the kubeconfig to a db
  const tmpClusterKubeconfig = tmp.fileSync();
  fs.writeFileSync(tmpClusterKubeconfig.name, clusterKubeconfig!);

  // set up the cluster!
  // if (
  //   (await exec(
  //     `KUBECONFIG=${tmpClusterKubeconfig.name} timoni bundle apply -f ../../timoni/bundles/cluster-necessities.cue`
  //   ).code) !== 0
  // ) {
  //   throw new Error(`Failed to apply cluster-necessities`);
  // }
  // if (
  //   (await exec(
  //     `KUBECONFIG=${tmpClusterKubeconfig.name} timoni bundle apply -f ../../timoni/bundles/flux-aio.cue`
  //   ).code) !== 0
  // ) {
  //   throw new Error(`Failed to apply flux-aio`);
  // }
  // if (
  //   (await exec(
  //     `KUBECONFIG=${tmpClusterKubeconfig.name} timoni bundle apply -f ../../timoni/bundles/cluster-addons.cue`
  //   ).code) !== 0
  // ) {
  //   throw new Error(`Failed to apply cluster-addons`);
  // }

  // cleanup
  mgmtKCFile.removeCallback();

  return new Response(clusterKubeconfig, {
    status: 200,
    headers: {
      "Content-Type": "application/yaml",
    },
  });
}
