import * as k8s from "@kubernetes/client-node";
import { serviceName } from "@lib/constants";
import { SpanStatusCode, trace } from "@opentelemetry/api";
import { ExecException, exec as execCB } from "child_process";
import fs from "fs";
import { dump } from "js-yaml";
import tmp from "tmp";
import util from "util";
const exec = util.promisify(execCB);

const requiredEnvVars = [
  "HETZNER_MANAGEMENT_CLUSTER_NAME",
  "HETZNER_MANAGEMENT_CLUSTER_SERVER",
  "HETZNER_MANAGEMENT_CLUSTER_CA_DATA",
  "HETZNER_MANAGEMENT_CLUSTER_USER_CERT_DATA",
  "HETZNER_MANAGEMENT_CLUSTER_USER_KEY_DATA",
];
for (const env of requiredEnvVars) {
  if (!process.env[env]) {
    throw new Error(`${env} is not set`);
  }
}

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
fs.closeSync(mgmtKCFile.fd);

export function mgmtClusterKubeconfigFile(): string {
  return mgmtKCFile.name;
}

// tracedExec is a function that runs a command and traces it
// it takes in the span name, attributes to set, and the command to run
// it returns true on success, false if the command fails
// it adds a lot of attributes in order to make debugging easier
export async function tracedExec({
  spanName,
  spanAttributes,
  command,
  directory,
}: {
  spanName: string;
  spanAttributes?: Record<string, string>;
  command: string;
  directory?: string;
}): Promise<{ stdout: string; stderr: string }> {
  return await trace
    .getTracer(serviceName)
    .startActiveSpan(spanName, async (span) => {
      if (spanAttributes) {
        span.setAttributes(spanAttributes);
      }
      try {
        console.log("DEBUG: exec", command);
        const ret = await exec(command, { cwd: directory });
        span.setAttributes(ret);
        span.end();
        return ret;
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
        throw new Error(`${spanName}: ${error.stdout} ${error.stderr}`);
      }
    });
}

export function generateAwsAccessKeyId() {
  const prefix = "AKIA";
  const length = 16; // Total length 20 - prefix length 4
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  let result = prefix;
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

export function generateAwsSecretAccessKey() {
  const length = 40;
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

type StringObject = { [key: string]: string };

export async function findOrCreateSecret<T extends StringObject>(
  kubeconfigFilename: string,
  namespace: string,
  secretName: string,
  defaultValues: T
): Promise<T> {
  try {
    const { stdout } = await exec(
      `kubectl --kubeconfig=${kubeconfigFilename} get secret ${secretName} -n ${namespace} -o json`
    );
    const existingSecret = JSON.parse(stdout);
    const data: StringObject = {};
    for (const key in defaultValues) {
      if (existingSecret.data?.[key] !== undefined) {
        data[key] = Buffer.from(
          existingSecret.data[key] || "",
          "base64"
        ).toString("utf-8");
      }
    }
    return data as T;
  } catch (e: any) {
    const error = e as ExecException;
    if (error.code === 1) {
      const literals = Object.entries(defaultValues)
        .map(([key, value]) => `--from-literal=${key}=${value}`)
        .join(" ");
      await exec(
        `kubectl --kubeconfig=${kubeconfigFilename} create secret generic ${secretName} -n ${namespace} ${literals}`
      );
      return defaultValues;
    } else {
      throw e;
    }
  }
}

export async function findOrCreateNamespace(
  kubeconfigFilename: string,
  namespace: string
): Promise<void> {
  try {
    await exec(
      `kubectl --kubeconfig=${kubeconfigFilename} get namespace ${namespace}`
    );
  } catch (e: any) {
    const error = e as ExecException;
    if (error.code === 1) {
      // Namespace not found, create it
      await exec(
        `kubectl --kubeconfig=${kubeconfigFilename} create namespace ${namespace}`
      );
    } else {
      throw e;
    }
  }
}
