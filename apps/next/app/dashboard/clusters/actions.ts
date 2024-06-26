"use server";
import { db } from "@/app/server/db";
import {
  HetznerCluster,
  HetznerProject,
  HetznerProjectSpec,
  Team,
  User,
  hetznerClusters,
  hetznerProjects,
  teams,
  users,
} from "@/app/server/db/schema";
import { getRunOutput, inngest } from "@/lib/inngest";
import { auth } from "@clerk/nextjs/server";
import { ServerActionState } from "@lib/action";
import { eq } from "drizzle-orm";
import { redirect } from "next/navigation";
import uuidBase62 from "uuid-base62";

export async function createHetznerProject(
  prevState: ServerActionState,
  formData: FormData
): Promise<ServerActionState> {
  const { userId, getToken, orgId } = auth();
  if (!userId) {
    return { isError: true, message: "You're not signed in." };
  } else if (!orgId) {
    return { isError: true, message: "You're not in an organization." };
  }
  const projectName: string = formData.get("projectName") as string;
  if (!projectName) {
    return { isError: true, message: "Project name is required." };
  }
  const apiKey = formData.get("apiKey") as string;
  if (!apiKey) {
    return { isError: true, message: "API key is required." };
  }
  const token = await getToken();
  if (!token) {
    return { isError: true, message: "You're not authorized." };
  }

  // pull the user from the database via userId (aka clerkId in our db)
  const user: User | undefined = await db
    .select()
    .from(users)
    .where(eq(users.clerkId, userId))
    .then((rows) => rows[0] || undefined);
  if (!user) {
    return {
      isError: true,
      message: "Couldn't find your user in the database.",
    };
  }
  const team: Team | undefined = await db
    .select()
    .from(teams)
    .where(eq(teams.clerkId, orgId))
    .then((rows) => rows[0] || undefined);
  if (!team) {
    return {
      isError: true,
      message: "Couldn't find your team in the database.",
    };
  }

  const id = uuidBase62.v4();
  const spec: HetznerProjectSpec = {
    id,
    teamId: team.id,
    hetznerName: projectName,
    hetznerApiToken: apiKey,
    creatorId: user.id,
  };
  let result: HetznerProject | undefined;
  const { ids: eventIds } = await inngest.send({
    name: "hetzner-project/create",
    data: spec,
  });
  const output = await getRunOutput(eventIds[0]!);
  if (output.type && output.message) {
    return {
      isError: true,
      message: `Hetzner connection failed: ${output.message}`,
    };
  }
  redirect(`/dashboard/clusters`);
}

export async function deleteHetznerCluster(
  prevState: ServerActionState,
  formData: FormData
): Promise<ServerActionState> {
  const clusterId: string = formData.get("clusterId") as string;
  const cluster: HetznerCluster | undefined = await db
    .select()
    .from(hetznerClusters)
    .where(eq(hetznerClusters.id, clusterId))
    .then((rows) => rows[0] || undefined);
  if (!cluster) {
    return { isError: true, message: "Cluster not found." };
  }
  // todo: figure out how to cancel createHetznerCluster and provisionHetznerCluster workflows, in case either is still running
  await inngest.send({
    name: "hetzner-cluster/delete",
    data: { clusterId },
  });
  return { isError: false, message: "Cluster deleted." };
}

export async function fetchProjectAndClusters(): Promise<{
  project: HetznerProject | undefined;
  clusters: HetznerCluster[];
}> {
  const { orgId } = auth();
  if (!orgId) {
    throw new Error("No orgId found");
  }
  const team: Team | undefined = await db
    .select()
    .from(teams)
    .where(eq(teams.clerkId, orgId))
    .then((rows) => rows[0] || undefined);
  if (!team) {
    throw new Error("No team found");
  }
  const project: HetznerProject | undefined = await db
    .select()
    .from(hetznerProjects)
    .where(eq(hetznerProjects.teamId, team.id))
    .then((rows) => rows[0] || undefined);
  const clusters: HetznerCluster[] = await db
    .select()
    .from(hetznerClusters)
    .where(eq(hetznerClusters.teamId, team.id));
  return { project, clusters };
}
