"use server";

import { db } from "@/app/server/db";
import { hetznerClusters } from "@/app/server/db/schema";
import dayjs from "dayjs";
import { eq } from "drizzle-orm";
import { ClusterData, TimeSeries } from "./data";

export async function fetchClusterMetrics({
  timeframeSeconds,
  clusterName,
}: {
  timeframeSeconds: number;
  clusterName: string;
}): Promise<{
  cpu: TimeSeries[];
  mem: TimeSeries[];
  cpuRequests: TimeSeries[];
  memRequests: TimeSeries[];
}> {
  const cluster = await db.query.hetznerClusters.findFirst({
    where: eq(hetznerClusters.name, clusterName),
  });
  if (!cluster) {
    throw new Error("Cluster not found");
  }
  const clusterData = new ClusterData(cluster);
  const now = dayjs();
  const range = {
    startDate: now.subtract(timeframeSeconds, "seconds").toDate(),
    endDate: now.toDate(),
  };
  // TODO: Need a better mechanism, e.g. exposing prometheus API directly w/ some HTTP basic auth (similar to docker)
  const cpu = await clusterData.cpu(range);
  const mem = await clusterData.mem(range);
  const cpuRequests = await clusterData.cpuRequests(range);
  const memRequests = await clusterData.memRequests(range);
  return { cpu, mem, cpuRequests, memRequests };
}
