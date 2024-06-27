import { db } from "@/app/server/db";
import { Topbar } from "./Topbar";
import { hetznerClusters } from "@/app/server/db/schema";
import { eq } from "drizzle-orm";
import { mustGetActiveTeam } from "@/app/server/user";
import Content from "./Content";
import { fetchClusterMetrics } from "./actions";
import { inngest } from "@/lib/inngest";

export default async function AsyncClusterDetail({
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
    <>
      <Topbar cluster={cluster} />
      <Content clusterName={cluster.name} initialData={initialData} />
    </>
  );
}
