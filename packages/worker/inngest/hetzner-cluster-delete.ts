import { db } from "@metal/webapp/app/server/db";
import type { HetznerCluster, Team } from "@metal/webapp/app/server/db/schema";
import { hetznerClusters, teams } from "@metal/webapp/app/server/db/schema";
import { serviceName } from "@metal/webapp/lib/constants";
import { tracedExec } from "@metal/webapp/lib/tracedExec";
import { trace } from "@opentelemetry/api";
import { eq } from "drizzle-orm";
import fs from "fs";
import tmp from "tmp";
import { inngest } from "./client";
import { mgmtClusterKubeconfigFile } from "./shared";

export const hetznerClusterDelete = inngest.createFunction(
  { id: "hetzner-cluster-delete" },
  { event: "hetzner-cluster/delete" },
  async ({ event, step }) => {
    return await trace
      .getTracer(serviceName)
      .startActiveSpan("deleteHetznerCluster", async (span) => {
        const clusterId = event.data.clusterId;
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

        // delete things that involve pvcs so that there are no orphaned pvcs
        if (cluster.kubeconfig) {
          const clusterKubeconfigFile = tmp.fileSync();
          fs.writeFileSync(clusterKubeconfigFile.name, cluster.kubeconfig);
          for (const command of [
            `KUBECONFIG=${clusterKubeconfigFile.name} helm uninstall kube-prometheus --cascade foreground -n monitoring`,
            `KUBECONFIG=${clusterKubeconfigFile.name} helm uninstall minio-tenant --cascade foreground -n minio-tenant`,
            `KUBECONFIG=${clusterKubeconfigFile.name} helm uninstall registry --cascade foreground -n registry`,
          ]) {
            await tracedExec({
              spanName: "exec-kubectl-delete-things-with-pvcs",
              spanAttributes,
              command,
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
);
