import { db } from "@metal/webapp/app/server/db";
import type {
  HetznerCluster,
  HetznerNodeGroup,
  HetznerProject,
  Team,
} from "@metal/webapp/app/server/db/schema";
import {
  hetznerClusters,
  hetznerNodeGroups,
  hetznerProjects,
  teams,
} from "@metal/webapp/app/server/db/schema";
import { serviceName } from "@metal/webapp/lib/constants";
import { findOrCreateNamespace } from "@metal/webapp/lib/k8s";
import { tracedExec } from "@metal/webapp/lib/tracedExec";
import { trace } from "@opentelemetry/api";
import { eq } from "drizzle-orm";
import fs from "fs";
import path from "path";
import tmp from "tmp";
import * as ub62 from "uuid-base62";
import { inngest } from "./client";
import {
  findOrCreateSecret,
  generateAwsAccessKeyId,
  generateAwsSecretAccessKey,
  mgmtClusterKubeconfigFile,
} from "./shared";

// teamHetznerSecretName is the name of the secret in the management cluster containing the team's hetzner creds.
function teamHetznerSecretName(teamId: string): string {
  const teamUuid = ub62.decode(teamId); // will use this in k8s secret name, which must be [a-z0-9-]+
  return `${teamUuid}-hetzner`;
}

async function helmReleaseExists(
  kubeconfigFilename: string,
  namespace: string,
  releaseName: string
): Promise<boolean> {
  try {
    const { stdout } = await tracedExec({
      spanName: "check-helm-release-status",
      spanAttributes: {},
      command: `KUBECONFIG=${kubeconfigFilename} helm status -n ${namespace} ${releaseName}`,
    });
    return stdout.includes("STATUS: deployed");
  } catch (e) {
    return false;
  }
}

// envString turns an object into the VAR=value string that can be used to prefix a command
function envString(env: Record<string, string | number>): string {
  return (
    Object.entries(env)
      .map(([key, value]) => `${key}=${value}`)
      .join(" ") + " "
  );
}

export const hetznerClusterProvision = inngest.createFunction(
  { id: "hetzner-cluster-provision" },
  { event: "hetzner-cluster/provision" },
  async ({ event, step }) => {
    return await trace
      .getTracer(serviceName)
      .startActiveSpan("provisionHetznerCluster", async (span) => {
        const clusterId = event.data.clusterId;
        const cluster: HetznerCluster | undefined =
          await db.query.hetznerClusters.findFirst({
            where: eq(hetznerClusters.id, clusterId),
          });
        if (!cluster) {
          throw new Error("Cluster not found");
        }
        const nodeGroups: HetznerNodeGroup[] =
          await db.query.hetznerNodeGroups.findMany({
            where: eq(hetznerNodeGroups.clusterId, clusterId),
          });
        if (nodeGroups.length === 0) {
          throw new Error("No node groups found");
        }
        const hetznerProject: HetznerProject | undefined =
          await db.query.hetznerProjects.findFirst({
            where: eq(hetznerProjects.teamId, cluster.teamId),
          });
        if (!hetznerProject) {
          throw new Error("Project not found");
        }
        const team: Team | undefined = await db.query.teams.findFirst({
          where: eq(teams.id, cluster.teamId),
        });
        if (!team) {
          throw new Error("Team not found");
        }
        const spanAttributes = {
          hetznerClusterId: clusterId,
          hetznerName: cluster.name,
          teamId: cluster.teamId,
          creatorId: cluster.creatorId,
        };
        span.setAttributes(spanAttributes);
        return {
          event,
          body: await _provisionHetznerK8sCluster({
            cluster,
            nodeGroups,
            team,
            hetznerProject,
            spanAttributes,
          }),
        };
      });
  }
);

// _provisionH8sCluster provisions a full-on k8s cluster using cluster api hetzner and kubeadm control plane / bootstrap provider
async function _provisionHetznerK8sCluster({
  cluster,
  nodeGroups,
  team,
  hetznerProject,
  spanAttributes,
}: {
  cluster: HetznerCluster;
  nodeGroups: HetznerNodeGroup[];
  team: Team;
  hetznerProject: HetznerProject;
  spanAttributes: Record<string, string>;
}) {
  const mgmtKubeconfig = mgmtClusterKubeconfigFile();
  let clusterKubeconfig = cluster.kubeconfig;
  if (!clusterKubeconfig || clusterKubeconfig?.length === 0) {
    // TODO: support > 1 node group
    const nodeGroup = nodeGroups[0];
    if (!nodeGroup) {
      throw new Error("No node groups found");
    }

    // first things first get the letsencrypt cert for the cluster cooking since it takes 1-2 mins
    const tmpDirForCert = tmp.dirSync();
    await fs.writeFileSync(
      path.join(tmpDirForCert.name, "cert.yaml"),
      `apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ${cluster.name}-certificate
  namespace: cert-manager
spec:
  dnsNames:
  - '*.${cluster.name}.up.onmetal.dev'
  issuerRef:
    kind: ClusterIssuer
    name: letsencrypt-production
  secretName: ${cluster.name}-certificate
`
    );
    await tracedExec({
      spanName: "setup-cert",
      spanAttributes,
      command: `KUBECONFIG=${mgmtKubeconfig} kubectl apply -f cert.yaml`,
      directory: tmpDirForCert.name,
    });

    // add project creds as secrets to the mgmt cluster.
    // hetzner cluster resources will reference this
    await findOrCreateSecret(
      mgmtKubeconfig,
      "default",
      teamHetznerSecretName(team.id),
      {
        hcloud: hetznerProject.hetznerApiToken,
        "robot-user": hetznerProject.hetznerWebserviceUsername!,
        "robot-password": hetznerProject.hetznerWebservicePassword!,
      }
    );

    let { stdout: clusterctlVersion } = await tracedExec({
      spanName: "clusterctl-version",
      spanAttributes,
      command: `clusterctl version -o short`,
    });
    clusterctlVersion = clusterctlVersion.trim();
    await db
      .update(hetznerClusters)
      .set({ clusterctlVersion })
      .where(eq(hetznerClusters.id, cluster.id));

    const tmpDirForClusterTemplate = tmp.dirSync();
    await tracedExec({
      spanName: "download-cluster-template",
      spanAttributes,
      command: `wget https://github.com/syself/cluster-api-provider-hetzner/releases/download/v1.0.0-beta.33/cluster-template-hcloud.yaml -O cluster-template-hcloud.yaml`,
      directory: tmpDirForClusterTemplate.name,
    });
    // use kustomize to do some modifications to the HetznerCluster:
    //  a) specify the hetznerSecretRef.name
    //  b) add a load balancer http and https services to allow for http[s]://*.<cluster name>.up.onmetal.dev requests to enter the cluster (will set up a gateway to route them further later on)
    fs.writeFileSync(
      path.join(tmpDirForClusterTemplate.name, "hetzner-cluster-patch.yaml"),
      `apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: HetznerCluster
metadata:
  name: \${CLUSTER_NAME}
spec:
  controlPlaneEndpoint:
    port: 6443
  controlPlaneLoadBalancer:
    enabled: true
    port: 6443
    # TODO: parameterize this in the future in case of scale up
    type: lb11
    extraServices:
    - protocol: tcp
      listenPort: 443
      destinationPort: 443
      # TODO: PR something on caph to enable proxy protocol (would send client info)
    - protocol: tcp
      listenPort: 80
      destinationPort: 80
  hetznerSecretRef:
    name: ${teamHetznerSecretName(team.id)}`
    );
    fs.writeFileSync(
      path.join(tmpDirForClusterTemplate.name, "kustomization.yaml"),
      `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-template-hcloud.yaml
patches:
- path: hetzner-cluster-patch.yaml
`
    );
    const { stdout: clusterTemplateKustomized } = await tracedExec({
      spanName: "kustomize-build",
      spanAttributes,
      command: `kustomize build`,
      directory: tmpDirForClusterTemplate.name,
    });
    fs.writeFileSync(
      path.join(
        tmpDirForClusterTemplate.name,
        "cluster-template-kustomized.yaml"
      ),
      clusterTemplateKustomized
    );

    const { stdout: clusterTemplate } = await tracedExec({
      spanName: "exec-clusterctl-generate-cluster",
      spanAttributes,
      command:
        envString({
          HCLOUD_CONTROL_PLANE_MACHINE_TYPE: nodeGroup?.instanceType,
          HCLOUD_REGION: cluster.location,
          HCLOUD_SSH_KEY: hetznerProject.sshKeyName!,
          HCLOUD_WORKER_MACHINE_TYPE: nodeGroup?.instanceType,
          KUBECONFIG: mgmtKubeconfig,
          KUBERNETES_VERSION: cluster.k8sVersion,
          CONTROL_PLANE_MACHINE_COUNT: 1,
          WORKER_MACHINE_COUNT: nodeGroup!.minNodes - 1,
        }) +
        `clusterctl generate cluster ${cluster.name} --from ./cluster-template-kustomized.yaml`,
      directory: tmpDirForClusterTemplate.name,
    });
    fs.writeFileSync(
      path.join(tmpDirForClusterTemplate.name, "cluster-template.yaml"),
      clusterTemplate
    );
    await db
      .update(hetznerClusters)
      .set({
        status: "initializing",
        clusterctlTemplate: clusterTemplate,
      })
      .where(eq(hetznerClusters.id, cluster.id));
    await tracedExec({
      spanName: "exec-kubectl-apply-cluster-config",
      spanAttributes,
      command: `KUBECONFIG=${mgmtKubeconfig} kubectl apply -f ${tmpDirForClusterTemplate.name}/cluster-template.yaml`,
      directory: tmpDirForClusterTemplate.name,
    });

    // wait for kubeconfig to become available, save it to the db
    const timeout = 20 * 60 * 1000;
    const start = Date.now();
    do {
      const { stdout } = await tracedExec({
        spanName: "check-for-control-plane-ready",
        spanAttributes,
        command: `KUBECONFIG=${mgmtKubeconfig} kubectl get kubeadmcontrolplane -ojson -l cluster.x-k8s.io/cluster-name=${cluster.name}`,
      });
      const kubeadmControlPlanes = JSON.parse(stdout);
      if (
        kubeadmControlPlanes.items.length === 1 &&
        kubeadmControlPlanes.items[0].status?.conditions &&
        kubeadmControlPlanes.items[0].status.conditions.some(
          (condition: { type: string; status: string }) =>
            condition.type === "CertificatesAvailable" &&
            condition.status === "True"
        )
      ) {
        const { stdout } = await tracedExec({
          spanName: "get-kubeconfig",
          spanAttributes,
          command: `KUBECONFIG=${mgmtKubeconfig} clusterctl get kubeconfig ${cluster.name}`,
        });
        clusterKubeconfig = stdout;
      }
      if (clusterKubeconfig === null) {
        await new Promise((resolve) => setTimeout(resolve, 5000));
      }
    } while (clusterKubeconfig === null && Date.now() - start < timeout);
    if (clusterKubeconfig === null) {
      throw new Error(`Timed out waiting for control plane to be ready`);
    }

    await db
      .update(hetznerClusters)
      .set({
        kubeconfig: clusterKubeconfig,
      })
      .where(eq(hetznerClusters.id, cluster.id));
  }

  // write clusterKubeconfig to tmp file for use by kubectl et al
  const clusterKubeconfigFile = tmp.fileSync();
  fs.writeFileSync(clusterKubeconfigFile.name, clusterKubeconfig);

  // wait for control plane to come up by looping on kubectl get --raw='/readyz?verbose' until it returns exit code 0
  let timeout = 20 * 60 * 1000;
  let start = Date.now();
  do {
    try {
      const { stdout } = await tracedExec({
        spanName: "check-for-control-plane-ready",
        spanAttributes,
        command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl get --raw='/readyz?verbose'`,
      });
      if (stdout.includes("readyz check passed")) {
        break;
      }
    } catch (e) {
      // ignore errors since the control plane may not be ready yet
    }
    await new Promise((resolve) => setTimeout(resolve, 5000));
  } while (Date.now() - start < timeout);

  // install cni
  const ciliumExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "kube-system",
    "cilium"
  );
  if (!ciliumExists) {
    await tracedExec({
      spanName: "install-cni-repo",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add cilium https://helm.cilium.io/`,
    });
    await tracedExec({
      spanName: "install-cni",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install cilium cilium/cilium \
        --version 1.15.4 \
        --namespace kube-system \
        --set image.pullPolicy=IfNotPresent \
        --set ipam.mode=kubernetes \
        --set gatewayAPI.enabled=true \
        --set nodePort.enabled=true \
        --set hubble.relay.enabled=true \
        --set hubble.ui.enabled=true \
        --set kubeProxyReplacement=true \
        --set l2announcements.enabled=true \
        --set k8sClientRateLimit.qps=100 \
        --set k8sClientRateLimit.burst=200 \
        --set rollOutCiliumPods=true \
        --set operator.rollOutPods=true \
        --set operator.replicas=1`, // todo: set this to two once cluster size > 1
    });
  }

  // install gateway API CRDs
  await tracedExec({
    spanName: "install-gateway-api",
    spanAttributes,
    command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/experimental-install.yaml`,
  });

  // wait for cilium to come up
  timeout = 20 * 60 * 1000;
  start = Date.now();
  let ciliumReady = false;
  do {
    const { stdout } = await tracedExec({
      spanName: "check-for-cilium-ready",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl get daemonset -n kube-system -ojson cilium`,
    });
    const daemonSet = JSON.parse(stdout);
    if (
      daemonSet.status?.numberReady &&
      daemonSet.status.numberReady === daemonSet.status.desiredNumberScheduled
    ) {
      ciliumReady = true;
    } else {
      await new Promise((resolve) => setTimeout(resolve, 5000));
    }
  } while (!ciliumReady && Date.now() - start < timeout);
  if (clusterKubeconfig === null) {
    throw new Error(`Timed out waiting for cilium to be ready`);
  }

  // install ccm
  const ccmExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "kube-system",
    "hccm"
  );
  if (!ccmExists) {
    // ccm requires a secret on the cluster named 'hcloud' that includes the network name in hetzner cloud
    await findOrCreateSecret(
      clusterKubeconfigFile.name,
      "kube-system",
      "hcloud",
      {
        token: hetznerProject.hetznerApiToken,
        "robot-user": hetznerProject.hetznerWebserviceUsername!,
        "robot-password": hetznerProject.hetznerWebservicePassword!,
        network: cluster.name, // this relies on caph's convention of naming network after the cluster
      }
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add hcloud https://charts.hetzner.cloud`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install hccm hcloud/hcloud-cloud-controller-manager --version 1.19.0 --namespace kube-system`,
    ]) {
      await tracedExec({
        spanName: "install-ccm-repo",
        spanAttributes,
        command,
      });
    }
  }

  // install csi
  const csiExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "kube-system",
    "hcloud-csi"
  );
  if (!csiExists) {
    await tracedExec({
      spanName: "install-csi",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install hcloud-csi hcloud/hcloud-csi --version 2.6.0 \
      --namespace kube-system`,
    });
  }

  // add l2announcmentpolicy and ip pool. These will let us use cilium as an l2 load balancer routing external traffic into the cluster
  timeout = 20 * 60 * 1000;
  start = Date.now();
  let externalIP = "";
  do {
    const { stdout } = await tracedExec({
      spanName: "get-external-ip",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl get node -ojson | jq -r '.items[0].status.addresses[] | select(.type == "ExternalIP").address'`,
    });
    const ip = stdout.trim();
    if (ip) {
      externalIP = ip;
      break;
    }
    await new Promise((resolve) => setTimeout(resolve, 5000));
  } while (Date.now() - start < timeout);
  const tmpDirForCilium = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForCilium.name, "l2announcementpolicy.yaml"),
    `apiVersion: cilium.io/v2alpha1
kind: CiliumL2AnnouncementPolicy
metadata:
  name: default-l2-announcement-policy
  namespace: kube-system
spec:
  interfaces:
  - ^eth[0-9]+
  loadBalancerIPs: true
  externalIPs: true`
  );
  await fs.writeFileSync(
    path.join(tmpDirForCilium.name, "ippool.yaml"),
    `apiVersion: cilium.io/v2alpha1
kind: CiliumLoadBalancerIPPool
metadata:
  name: default-pool
  namespace: kube-system
spec:
  cidrs:
    - cidr: "${externalIP}/32"`
  );
  for (const command of [
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -f l2announcementpolicy.yaml`,
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -f ippool.yaml`,
  ]) {
    await tracedExec({
      spanName: "configure-cilium-l2",
      spanAttributes,
      command,
      directory: tmpDirForCilium.name,
    });
  }

  // set up dns in the mgmt cluster to point *.<cluster name>.up.onmetal.dev at the cluster LB
  // TODO: this actually points at the instance, not the LB. Fix that
  const tmpDirForDNSAndCert = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForDNSAndCert.name, "dns.yaml"),
    `apiVersion: externaldns.k8s.io/v1alpha1
kind: DNSEndpoint
metadata:
  name: ${cluster.name}
  namespace: external-dns
spec:
  endpoints:
  - dnsName: "*.${cluster.name}.up.onmetal.dev"
    recordType: A
    targets:
    - ${externalIP}`
  );
  await tracedExec({
    spanName: "setup-dns-and-cert",
    spanAttributes,
    command: `KUBECONFIG=${mgmtKubeconfig} kubectl apply -f dns.yaml`,
    directory: tmpDirForDNSAndCert.name,
  });

  // remove the noschedule taint on the control plane node(s)
  // can add this back once the cluster grows to be big enough
  try {
    await tracedExec({
      spanName: "remove-noschedule-taint",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl taint nodes --all node-role.kubernetes.io/control-plane-`,
    });
  } catch (e) {
    // ignore, might not exist
  }

  // install metrics-server
  const metricsServerExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "kube-system",
    "metrics-server"
  );
  if (!metricsServerExists) {
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install metrics-server metrics-server/metrics-server --version 3.12.1 --namespace kube-system`,
    ]) {
      await tracedExec({
        spanName: "install-metrics-server",
        spanAttributes,
        command,
      });
    }
  }

  // install prometheus
  const prometheusExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "monitoring",
    "kube-prometheus"
  );
  let prometheusPromise:
    | Promise<{ stdout: string; stderr: string }>
    | undefined = undefined;
  if (!prometheusExists) {
    const tmpDirForPrometheus = tmp.dirSync();
    await fs.writeFileSync(
      path.join(tmpDirForPrometheus.name, "prometheus-values.yaml"),
      `
grafana:
  # enable persistence so we can do things like install plugins w/o a restart killing the config
  persistence:
    enabled: true
    type: sts
    storageClassName: "hcloud-volumes"
    accessModes:
      - ReadWriteOnce
    size: 20Gi
    finalizers:
      - kubernetes.io/pvc-protection
  # add quickwit and jaeger datasources, which will be created later on
  additionalDataSources:
  - name: quickwit-logs
    type: quickwit-quickwit-datasource
    orgId: 1
    access: proxy
    url: http://quickwit-searcher.monitoring.svc.cluster.local:7280/api/v1
    jsonData:
      index: "otel-logs-v0_7"
      logLevelField: ""
      logMessageField: "body.message"
  - name: jaeger-traces
    type: jaeger
    orgId: 1
    access: proxy
    url: http://jaeger-query.monitoring.svc.cluster.local:80
`
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add prometheus-community https://prometheus-community.github.io/helm-charts`,
    ]) {
      await tracedExec({
        spanName: "install-prometheus",
        spanAttributes,
        command,
        directory: tmpDirForPrometheus.name,
      });
    }
    // put this in the background bc it takes some time
    prometheusPromise = tracedExec({
      spanName: "install-prometheus",
      spanAttributes,
      command: `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install kube-prometheus prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace --version 58.4.0 --values prometheus-values.yaml`,
      directory: tmpDirForPrometheus.name,
    });
  }

  // copy the secret containing the cert over to the user's cluster
  await tracedExec({
    spanName: "wait-for-cert-to-be-ready",
    spanAttributes,
    command: `KUBECONFIG=${mgmtKubeconfig} kubectl wait --for=condition=ready --timeout=60s certificate -n cert-manager ${cluster.name}-certificate`,
  });
  const { stdout: secretData } = await tracedExec({
    spanName: "get-secret-data",
    spanAttributes,
    command: `KUBECONFIG=${mgmtKubeconfig} kubectl get secret ${cluster.name}-certificate --namespace=cert-manager -oyaml | yq '.data' -o json | jq -r -c .`,
  });
  const tmpDirForCertCopy = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForCertCopy.name, "cert-secret.yaml"),
    `---
apiVersion: v1
kind: Namespace
metadata:
  name: gateway
---
apiVersion: v1
type: kubernetes.io/tls
kind: Secret
metadata:
  name: ${cluster.name}-certificate
  namespace: gateway
data: ${secretData.trim()}
`
  );
  await tracedExec({
    spanName: "copy-cert-to-user-cluster",
    spanAttributes,
    command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -f cert-secret.yaml`,
    directory: tmpDirForCertCopy.name,
  });

  // set up gateway that will receive external traffic
  // useful links:
  // - https://isovalent.com/blog/post/tutorial-getting-started-with-the-cilium-gateway-api/
  // - https://blog.stonegarden.dev/articles/2023/12/cilium-gateway-api/
  // - https://docs.cilium.io/en/stable/network/servicemesh/gateway-api/https/
  const tmpDirForGateway = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForGateway.name, "gateway.yaml"),
    `---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: cilium
  namespace: gateway
spec:
  controllerName: io.cilium/gateway-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: cilium
  namespace: gateway
spec:
  gatewayClassName: cilium
  infrastructure:
    annotations:
      io.cilium/lb-ipam-ips: "${externalIP}"
  listeners:
    - protocol: HTTP
      port: 80
      name: http-subdomains-gateway
      hostname: "*.${cluster.name}.up.onmetal.dev"
      allowedRoutes:
        namespaces:
          from: All
    - protocol: HTTPS
      port: 443
      name: https-subdomains-gateway
      hostname: "*.${cluster.name}.up.onmetal.dev"
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: ${cluster.name}-certificate
            namespace: gateway
      allowedRoutes:
        namespaces:
          from: All`
  );
  for (const command of [
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n gateway -f gateway.yaml`,
  ]) {
    await tracedExec({
      spanName: "install-gateway",
      spanAttributes,
      command,
      directory: tmpDirForGateway.name,
    });
  }

  // minio for s3-compatible storage
  const minioOperatorExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "minio-operator",
    "minio-operator"
  );
  const minioTenantExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "minio-tenant",
    "minio-tenant"
  );
  await findOrCreateNamespace(clusterKubeconfigFile.name, "minio-tenant");
  const minioRootCreds = await findOrCreateSecret(
    clusterKubeconfigFile.name,
    "minio-tenant",
    "minio-tenant-root-credentials",
    {
      user: generateAwsAccessKeyId(),
      password: generateAwsSecretAccessKey(),
    }
  );
  if (!minioOperatorExists || !minioTenantExists) {
    const tmpDirForMinio = tmp.dirSync();
    await fs.writeFileSync(
      path.join(tmpDirForMinio.name, "minio-tenant-configuration.yaml"),
      `
apiVersion: v1
kind: Secret
metadata:
  name: minio-tenant-configuration
  namespace: minio-tenant
type: Opaque
stringData:
  config.env: |-
    export MINIO_ROOT_USER=${minioRootCreds.user}
    export MINIO_ROOT_PASSWORD=${minioRootCreds.password}
`
    );
    await fs.writeFileSync(
      path.join(tmpDirForMinio.name, "minio-tenant-values.yaml"),
      `
secrets: null
tenant:
  name: minio-tenant
  configuration:
    name: minio-tenant-configuration
  configSecret:
    name: minio-tenant-configuration
    existingSecret: true
  certificate:
    requestAutoCert: false
  pools:
  - servers: 1
    name: pool-0
    volumesPerServer: 1
    size: 50Gi
    storageClassName: hcloud-volumes`
    );
    await fs.writeFileSync(
      path.join(tmpDirForMinio.name, "minio-tenant-httproute.yaml"),
      `kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: minio-tenant
  namespace: minio-tenant
spec:
  parentRefs:
  - kind: Gateway
    name: cilium
    namespace: gateway
    port: 443
  hostnames:
  - 's3.${cluster.name}.up.onmetal.dev'
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: minio
      namespace: minio-tenant
      kind: Service
      port: 80`
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add minio-operator https://operator.min.io`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace minio-operator --create-namespace --version 5.0.15 minio-operator minio-operator/operator`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl wait --for=condition=available --timeout=60s deployments -n minio-operator minio-operator`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n minio-tenant -f minio-tenant-configuration.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace minio-tenant --version 5.0.15 minio-tenant minio-operator/tenant --values minio-tenant-values.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n minio-tenant -f minio-tenant-httproute.yaml`,
    ]) {
      await tracedExec({
        spanName: "install-minio",
        spanAttributes,
        command,
        directory: tmpDirForMinio.name,
      });
    }
    timeout = 20 * 60 * 1000;
    start = Date.now();
    do {
      try {
        const { stdout } = await tracedExec({
          spanName: "check-for-minio-pool-pod",
          spanAttributes,
          command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl get -n minio-tenant -l app=minio pod`,
        });
        if (stdout.includes("Running")) {
          break;
        }
      } catch (e) {
        // ignore errors since the pod might not be created
      }
      await new Promise((resolve) => setTimeout(resolve, 5000));
    } while (Date.now() - start < timeout);
  }

  // install quickwit + opentelemetry for logs and traces
  const quickwitExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "monitoring",
    "quickwit"
  );
  const opentelemetryExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "monitoring",
    "opentelemetry"
  );
  if (!quickwitExists || !opentelemetryExists) {
    // create access credentials for it to access the minio s3 endpoint
    const minioQuickwitCreds = await findOrCreateSecret(
      clusterKubeconfigFile.name,
      "monitoring",
      "quickwit-minio-access-credentials",
      {
        user: generateAwsAccessKeyId(),
        password: generateAwsSecretAccessKey(),
      }
    );
    const tmpDirForQuickwit = tmp.dirSync();
    await fs.writeFileSync(
      path.join(tmpDirForQuickwit.name, "quickwit-values.yaml"),
      `
environment:
  QW_METASTORE_URI: "s3://quickwit/quickwit-indexes"
  # this lets jaeger connect to quickwit over grpc
  QW_ENABLE_JAEGER_ENDPOINT: "true"
config:
  metastore_uri: "s3://quickwit/quickwit-indexes"
  default_index_root_uri: "s3://quickwit/quickwit-indexes"
  storage:
    s3:
      endpoint: "http://minio.minio-tenant.svc.cluster.local"
      region: "us-east-1"
      access_key_id: "${minioQuickwitCreds.user}"
      secret_access_key: "${minioQuickwitCreds.password}"
      flavor: minio
  indexer:
    # this lets quickwit receive grpc requests from otel collectors
    enable_otlp_endpoint: true
`
    );
    await fs.writeFileSync(
      path.join(tmpDirForQuickwit.name, "otel-values.yaml"),
      `
image:
  repository: "otel/opentelemetry-collector-k8s"
mode: daemonset
presets:
  logsCollection:
    enabled: true
  kubernetesAttributes:
    enabled: true
  kubernetesEvents:
    enabled: true
config:
  exporters:
    otlp/quickwit:
      endpoint: "quickwit-indexer.monitoring.svc.cluster.local:7281"
      tls:
        insecure: true
  service:
    pipelines:
      logs:
        exporters: ["otlp/quickwit"]
      traces:
        exporters: ["otlp/quickwit"]`
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl exec -n minio-tenant svc/minio -- /bin/bash -c "mc config host add ${cluster.name} http://minio.minio-tenant.svc.cluster.local:80 ${minioRootCreds.user} ${minioRootCreds.password} && mc mb --ignore-existing ${cluster.name}/quickwit && mc admin user add ${cluster.name} ${minioQuickwitCreds.user} ${minioQuickwitCreds.password} && mc admin policy attach ${cluster.name} readwrite --user ${minioQuickwitCreds.user}"`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add quickwit https://helm.quickwit.io`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace monitoring --version 0.5.15 quickwit quickwit/quickwit --values quickwit-values.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add opentelemetry https://open-telemetry.github.io/opentelemetry-helm-charts`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace monitoring --version 0.90.1 opentelemetry opentelemetry/opentelemetry-collector --values otel-values.yaml`,
    ]) {
      await tracedExec({
        spanName: "install-quickwit",
        spanAttributes,
        command,
        directory: tmpDirForQuickwit.name,
      });
    }
  }

  // set up jaeger plugged in to quickwit
  const jaegerExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "monitoring",
    "jaeger"
  );
  if (!jaegerExists) {
    // helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
    const tmpDirForJaeger = tmp.dirSync();
    await fs.writeFileSync(
      path.join(tmpDirForJaeger.name, "jaeger-values.yaml"),
      `
provisionDataStore:
  cassandra: false
  elasticsearch: false
  kafka: false
storage:
  type: grpc-plugin
  grpcPlugin:
    extraEnv:
    - name: SPAN_STORAGE_TYPE
      value: grpc-plugin
    - name: GRPC_STORAGE_SERVER
      value: quickwit-searcher.monitoring.svc.cluster.local:7281
    - name: GRPC_STORAGE_TLS
      value: "false"
agent:
  enabled: false
collector:
  enabled: false
query:
  enabled: true
  agentSidecar:
    enabled: false
`
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add jaegertracing https://jaegertracing.github.io/helm-charts`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace monitoring --version 3.0.7 jaeger jaegertracing/jaeger --values jaeger-values.yaml`,
    ]) {
      await tracedExec({
        spanName: "install-jaeger",
        spanAttributes,
        command,
        directory: tmpDirForJaeger.name,
      });
    }
  }

  // set up and expose a docker registry
  const dockerRegistryExists = await helmReleaseExists(
    clusterKubeconfigFile.name,
    "registry",
    "registry"
  );
  if (!dockerRegistryExists) {
    await findOrCreateNamespace(clusterKubeconfigFile.name, "registry");
    const registryHttpBasicCreds = await findOrCreateSecret(
      clusterKubeconfigFile.name,
      "registry",
      "registry-http-basic-credentials",
      {
        user: generateAwsAccessKeyId(),
        password: generateAwsSecretAccessKey(),
      }
    );
    const tmpDirForRegistry = tmp.dirSync();
    await tracedExec({
      spanName: "htpasswd-for-registry",
      spanAttributes,
      command: `htpasswd -Bbc ./auth ${registryHttpBasicCreds.user} ${registryHttpBasicCreds.password}`,
      directory: tmpDirForRegistry.name,
    });
    const htpasswd = fs
      .readFileSync(path.join(tmpDirForRegistry.name, "auth"), "utf8")
      .trim();

    const registryConfig = `
version: 0.1
http:
  secret: hownowbrowncow
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
auth:
  htpasswd:
    realm: basic-realm
    path: /etc/docker/auth/htpasswd
log:
  level: debug
  fields:
    service: registry
storage:
  filesystem:
    rootdirectory: /var/lib/registry
  delete:
    enabled: true
  maintenance:
    uploadpurging:
      enabled: true
      age: 168h
      interval: 24h
      dryrun: false
    readonly:
      enabled: false
`;
    await fs.writeFileSync(
      path.join(tmpDirForRegistry.name, "docker-registry-config.yaml"),
      `
kind: Secret
apiVersion: v1
metadata:
  name: docker-registry-config
  namespace: registry
data:
  config.yml: ${Buffer.from(registryConfig).toString("base64")}
`
    );
    await fs.writeFileSync(
      path.join(tmpDirForRegistry.name, "docker-registry-auth.yaml"),
      `
kind: Secret
apiVersion: v1
metadata:
  name: docker-registry-auth
  namespace: registry
data:
  htpasswd: ${Buffer.from(htpasswd).toString("base64")}
`
    );
    await fs.writeFileSync(
      path.join(tmpDirForRegistry.name, "registry-values.yaml"),
      `
fullnameOverride: docker-registry
ui:
  enabled: false
redis:
  enabled: false
storj:
  enabled: false
externalConfig:
  secretRef:
    name: docker-registry-config
extraVolumeMounts:
- name: auth
  mountPath: /etc/docker/auth
  readOnly: true
- name: data
  mountPath: /var/lib/registry/
extraVolumes:
- name: auth
  secret:
    secretName: docker-registry-auth
- name: data
  persistentVolumeClaim:
    claimName: docker-registry-pvc
`
    );
    fs.writeFileSync(
      path.join(tmpDirForRegistry.name, "docker-registry-pvc.yaml"),
      `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: docker-registry-pvc
  namespace: registry
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: hcloud-volumes`
    );
    await fs.writeFileSync(
      path.join(tmpDirForRegistry.name, "httproute.yaml"),
      `kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: docker-registry
  namespace: registry
spec:
  parentRefs:
  - kind: Gateway
    name: cilium
    namespace: gateway
    port: 443
  hostnames:
  - 'registry.${cluster.name}.up.onmetal.dev'
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: docker-registry
      kind: Service
      port: 5000`
    );
    for (const command of [
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl create -n registry -f docker-registry-pvc.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n registry -f docker-registry-config.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n registry -f docker-registry-auth.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm repo add mya https://mya.sh`,
      `KUBECONFIG=${clusterKubeconfigFile.name} helm upgrade --install --namespace registry --version 22.4.11 registry mya/registry --values registry-values.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n registry -f httproute.yaml`,
      `KUBECONFIG=${clusterKubeconfigFile.name} kubectl create secret docker-registry regcred-${cluster.name} --docker-server=registry.${cluster.name}.up.onmetal.dev --docker-username=${registryHttpBasicCreds.user} --docker-password=${registryHttpBasicCreds.password} --docker-email="doesntmatter@onmetal.dev"`,
    ]) {
      await tracedExec({
        spanName: "install-docker-registry",
        spanAttributes,
        command,
        directory: tmpDirForRegistry.name,
      });
    }
  }

  // argo rollouts https://github.com/argoproj-labs/rollouts-plugin-trafficrouter-gatewayapi/tree/main/examples/cilium
  // - https://github.com/argoproj-labs/rollouts-plugin-trafficrouter-gatewayapi/tree/main/examples/cilium
  // - https://rollouts-plugin-trafficrouter-gatewayapi.readthedocs.io/en/latest/quick-start/
  await findOrCreateNamespace(clusterKubeconfigFile.name, "argo-rollouts");
  for (const command of [
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml`,
  ]) {
    await tracedExec({
      spanName: "install-argo-rollouts",
      spanAttributes,
      command,
    });
  }

  // wait for argo-rollouts deployment to be in ready state and then install the argo gateway api plugin
  await tracedExec({
    spanName: "wait-for-argo-rollouts",
    spanAttributes,
    command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl wait --for=condition=available --timeout=60s deployments -n argo-rollouts argo-rollouts`,
  });
  const tmpDirForArgoGatewayApiPlugin = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForArgoGatewayApiPlugin.name, "argo.yaml"),
    `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argo-rollouts-config
  namespace: argo-rollouts
data:
  trafficRouterPlugins: |-
    - name: "argoproj-labs/gatewayAPI"
      location: "https://github.com/argoproj-labs/rollouts-plugin-trafficrouter-gatewayapi/releases/download/v0.3.0/gateway-api-plugin-linux-arm64"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gateway-controller-role
  namespace: argo-rollouts
rules:
  - apiGroups:
      - "*"
    resources:
      - "*"
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gateway-admin
  namespace: argo-rollouts
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: gateway-controller-role
subjects:
  - namespace: argo-rollouts
    kind: ServiceAccount
    name: argo-rollouts
`
  );
  for (const command of [
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -n argo-rollouts -f argo.yaml`,
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl rollout restart deployment/argo-rollouts -n argo-rollouts`, // restart necessary to pick up plugin configmap
  ]) {
    await tracedExec({
      spanName: "install-argo-gateway-api-plugin",
      spanAttributes,
      command,
      directory: tmpDirForArgoGatewayApiPlugin.name,
    });
  }

  // once prometheus install is done, add the quickwit plugin to grafana, and restart it for the plugin to get picked up
  if (prometheusPromise) {
    await prometheusPromise;
  }
  for (const command of [
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl exec -n monitoring svc/kube-prometheus-grafana -- grafana cli plugins install quickwit-quickwit-datasource 0.4.5`,
    `KUBECONFIG=${clusterKubeconfigFile.name} kubectl rollout restart -n monitoring statefulset/kube-prometheus-grafana`,
  ]) {
    await tracedExec({
      spanName: "configure-grafana",
      spanAttributes,
      command,
    });
  }

  // add an example rollout of a simple app (podinfo) that produces logs and traces
  await findOrCreateNamespace(clusterKubeconfigFile.name, "test-service");
  const tmpDirForTestService = tmp.dirSync();
  await fs.writeFileSync(
    path.join(tmpDirForTestService.name, "test-service.yaml"),
    `---
apiVersion: v1
kind: Service
metadata:
  name: test-service-stable
  namespace: test-service
spec:
  ports:
  - port: 80
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: test-service
---
apiVersion: v1
kind: Service
metadata:
  name: test-service-canary
  namespace: test-service
spec:
  ports:
  - port: 80
    targetPort: http
    protocol: TCP
    name: http
  selector:
    app: test-service
---
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: test-service-rollout
  namespace: test-service
spec:
  parentRefs:
  - kind: Gateway
    name: cilium
    namespace: gateway
    port: 443
  hostnames:
  - 'test-service.${cluster.name}.up.onmetal.dev'
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: test-service-stable
      kind: Service
      port: 80
    - name: test-service-canary
      kind: Service
      port: 80
---
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: test-service
  namespace: test-service
spec:
  replicas: 2
  strategy:
    canary:
      canaryService: test-service-canary
      stableService: test-service-stable
      trafficRouting:
        plugins:
          argoproj-labs/gatewayAPI:
            httpRoute: test-service-rollout-http-route
            namespace: test-service
      steps:
      - setWeight: 50
      - pause: {}
      - setWeight: 100
      - pause: {}
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      app: test-service
  template:
    metadata:
      labels:
        app: test-service
    spec:
      containers:
      - name: test-service
        image: stefanprodan/podinfo:6.6.2
        command: ['./podinfo', '--port', '9898', '--otel-service-name=podinfo']
        env:
        - name: NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: http://$(NODE_IP):4317
        ports:
        - name: http
          containerPort: 9898
          protocol: TCP
        resources:
          requests:
            memory: 32Mi
            cpu: 5m
`
  );
  await tracedExec({
    spanName: "install-test-service",
    spanAttributes,
    command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl apply -f test-service.yaml`,
    directory: tmpDirForTestService.name,
  });

  console.log("DONE");
  await db
    .update(hetznerClusters)
    .set({
      kubeconfig: clusterKubeconfig,
      status: "running",
    })
    .where(eq(hetznerClusters.id, cluster.id));
}
