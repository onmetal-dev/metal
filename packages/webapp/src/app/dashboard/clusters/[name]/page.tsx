import { db } from "@/app/server/db";
import { Topbar } from "./Topbar";
import { hetznerClusters } from "@/app/server/db/schema";
import { eq } from "drizzle-orm";
import { mustGetActiveTeam } from "@/app/server/user";
import Content from "./Content";
import { fetchClusterMetrics } from "./actions";
import { ContentLayout } from "@/components/dashboard/ContentLayout";

export default async function ClusterDetail({
  params,
}: {
  params: { name: string };
}) {
  const cluster = await db.query.hetznerClusters.findFirst({
    where: eq(hetznerClusters.name, params.name),
  });
  if (!cluster) {
    return <div>404</div>;
  }
  const team = await mustGetActiveTeam();
  if (team.id !== cluster.teamId) {
    return <div>404</div>;
  }

  const initialData = await fetchClusterMetrics({
    timeframeSeconds: 60 * 60,
    clusterName: cluster.name,
  });
  return (
    <ContentLayout title={cluster.name}>
      <Content clusterName={cluster.name} initialData={initialData} />
    </ContentLayout>
  );
}
