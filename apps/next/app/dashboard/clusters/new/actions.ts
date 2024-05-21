"use server";
import { auth } from "@clerk/nextjs";
import uuidBase62 from "uuid-base62";
import { createHetznerClusterState } from "./shared";
import {
  HetznerClusterInsert,
  HetznerInstanceTypeEnum,
  HetznerLocationEnum,
  HetznerNetworkZoneEnum,
  HetznerNodeGroupInsert,
  HetznerProject,
  Team,
  User,
  hetznerClusters,
  hetznerNodeGroups,
  hetznerProjects,
  teams,
  users,
} from "@/app/server/db/schema";
import { db } from "@/app/server/db";
import { eq } from "drizzle-orm";
import {
  uniqueNamesGenerator,
  adjectives,
  colors,
  animals,
} from "@joaomoreno/unique-names-generator";
import hetznerLocations from "@/lib/hcloud/locations";
import hetznerServerTypes from "@/lib/hcloud/server_types";
import { createTemporalClient } from "@/lib/temporal-client";
import { ProvisionHetznerCluster } from "@/temporal/src/workflows";
import { queueNameForEnv } from "@/lib/constants";
import { redirect } from "next/navigation";
import { networkZoneForLocation } from "@/lib/hcloud-helpers";

export async function createHetznerCluster(
  prevState: createHetznerClusterState,
  formData: FormData
): Promise<createHetznerClusterState> {
  const { userId, getToken, orgId } = auth();
  if (!userId) {
    return { isError: true, message: "You're not signed in." };
  } else if (!orgId) {
    return { isError: true, message: "You're not in an organization." };
  }

  // pull info out of FormData and validate it
  const datacenter: string = formData.get("datacenter")!.toString();
  if (
    !hetznerLocations.locations.find((location) => location.name === datacenter)
  ) {
    return { isError: true, message: "Invalid datacenter." };
  }
  const serverType: string = formData.get("serverType")!.toString();
  if (!hetznerServerTypes.server_types.find((st) => st.name === serverType)) {
    return { isError: true, message: "Invalid server type." };
  }
  const clusterSize: string = formData.get("clusterSize")!.toString();
  const clusterSizeInt = parseInt(clusterSize);
  if (isNaN(clusterSizeInt) || clusterSizeInt < 1 || clusterSizeInt > 100) {
    return { isError: true, message: "Invalid cluster size." };
  }

  // pull the user from the database via userId (aka clerkId in our db)
  const user: User | undefined = await db.query.users.findFirst({
    where: eq(users.clerkId, userId),
  });
  if (!user) {
    return {
      isError: true,
      message: "Couldn't find your user in the database.",
    };
  }
  const team: Team | undefined = await db.query.teams.findFirst({
    where: eq(teams.clerkId, orgId),
  });
  if (!team) {
    return {
      isError: true,
      message: "Couldn't find your team in the database.",
    };
  }
  const hetznerProject: HetznerProject | undefined =
    await db.query.hetznerProjects.findFirst({
      where: eq(hetznerProjects.teamId, team.id),
    });
  if (!hetznerProject) {
    return {
      isError: true,
      message: "Couldn't find your team's Hetzner project in the database.",
    };
  }

  const id = uuidBase62.v4();
  const cluster: HetznerClusterInsert = {
    id,
    creatorId: user.id,
    teamId: team.id,
    name: uniqueNamesGenerator({
      dictionaries: [adjectives, colors, animals],
      separator: "-",
      length: 2,
    }),
    status: "creating",
    k8sVersion: "1.28.8", // have to keep this in sync with what caph supports: https://github.com/syself/cluster-api-provider-hetzner. Also need to make sure ubuntu has it, e.g. `apt-get install -y kubelet=1.28.8-1.1`
    hetznerProjectId: hetznerProject.id,
    location: datacenter as HetznerLocationEnum,
    networkZone: networkZoneForLocation(datacenter) as HetznerNetworkZoneEnum,
  };
  const insertResult = await db
    .insert(hetznerClusters)
    .values(cluster)
    .returning({ insertedId: hetznerClusters.id });
  if (insertResult.length !== 1) {
    return {
      isError: true,
      message: "Failed to insert new cluster into database.",
    };
  }
  const clusterId = insertResult[0]!.insertedId;
  const nodeGroups: HetznerNodeGroupInsert[] = [
    {
      id: uuidBase62.v4(),
      clusterId,
      type: "all",
      instanceType: serverType as HetznerInstanceTypeEnum,
      minNodes: clusterSizeInt,
      maxNodes: clusterSizeInt,
    },
  ];
  await db.insert(hetznerNodeGroups).values(nodeGroups);

  const temporalClient = await createTemporalClient;
  // don't await the provision workflow since this does the bulk of the work and can take very long
  temporalClient.workflow.start(ProvisionHetznerCluster, {
    workflowId: `provisionHetznerCluster-${cluster.name}`,
    taskQueue: queueNameForEnv(process.env.NODE_ENV!),
    args: [{ clusterId }],
  });
  redirect(`/dashboard/clusters`);
}
