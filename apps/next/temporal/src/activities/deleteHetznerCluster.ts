import { db } from "@db/index";
import { HetznerCluster, Team, hetznerClusters, teams } from "@db/schema";
import { serviceName } from "@lib/constants";
import { trace } from "@opentelemetry/api";
import { eq } from "drizzle-orm";
import fs from "fs";
import tmp from "tmp";
import { mgmtClusterKubeconfigFile, tracedExec } from "./shared";

export async function deleteHetznerCluster({
  clusterId,
}: {
  clusterId: string;
}): Promise<void> {
  return await trace
    .getTracer(serviceName)
    .startActiveSpan("deleteHetznerCluster", async (span) => {
      const cluster: HetznerCluster | undefined = await db
        .select()
        .from(hetznerClusters)
        .where(eq(hetznerClusters.id, clusterId))
        .then((result) => result[0] || undefined);
      if (!cluster) {
        throw new Error("Cluster not found");
      }
      if (!cluster.clusterctlTemplate) {
        throw new Error("Cluster has now clusterctl template");
      }
      const team: Team | undefined = await db
        .select()
        .from(teams)
        .where(eq(teams.id, cluster.teamId))
        .then((result) => result[0] || undefined);
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

      // delete pvcs first so that there are no orphaned pvcs
      if (cluster.kubeconfig) {
        const clusterKubeconfigFile = tmp.fileSync();
        fs.writeFileSync(clusterKubeconfigFile.name, cluster.kubeconfig);
        const namespacesWithPvcs = ["monitoring", "minio-tenant", "registry"];
        for (const namespace of namespacesWithPvcs) {
          await tracedExec({
            spanName: "exec-kubectl-delete-pvcs",
            spanAttributes,
            command: `KUBECONFIG=${clusterKubeconfigFile.name} kubectl delete pvc -n ${namespace} --all`,
          });
        }
      }

      const mgmtKubeconfig = mgmtClusterKubeconfigFile();
      for (const command of [
        `KUBECONFIG=${mgmtKubeconfig} kubectl delete --ignore-not-found=true cluster ${cluster.name}`,
        `KUBECONFIG=${mgmtKubeconfig} kubectl delete --ignore-not-found=true -n external-dns dnsendpoint ${cluster.name}`, // todo: use ownerReferences to get these other deletes to happen?
        `KUBECONFIG=${mgmtKubeconfig} kubectl delete --ignore-not-found=true -n cert-manager certificate ${cluster.name}-certificate`,
        `KUBECONFIG=${mgmtKubeconfig} kubectl delete --ignore-not-found=true -n cert-manager secret ${cluster.name}-certificate`,
      ]) {
        tracedExec({
          spanName: "exec-kubectl-delete-cluster-template",
          spanAttributes,
          command,
        });
      }
      await db
        .update(hetznerClusters)
        .set({ status: "destroying" })
        .where(eq(hetznerClusters.id, cluster.id));
    });
}
