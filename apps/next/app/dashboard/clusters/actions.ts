"use server";
import { auth } from "@clerk/nextjs/server";
import uuidBase62 from "uuid-base62";
import { serverActionState } from "./shared";
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
import { eq } from "drizzle-orm";
import { createTemporalClient } from "@/lib/temporal-client";
import {
  CreateHetznerProject,
  DeleteHetznerCluster,
} from "@/temporal/src/workflows";
import { queueNameForEnv } from "@/lib/constants";
import {
  ApplicationFailure,
  WorkflowFailedError,
  WorkflowNotFoundError,
} from "@temporalio/client";
import { redirect } from "next/navigation";

export async function createHetznerProject(
  prevState: serverActionState,
  formData: FormData
): Promise<serverActionState> {
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
  const temporalClient = await createTemporalClient;
  let result: HetznerProject | undefined;
  try {
    const workflow = await temporalClient.workflow.start(CreateHetznerProject, {
      workflowId: `createHetznerProject-${spec.hetznerName}-${team.id}`, // for idempotency adopt a unique-enough convention on the workflow id
      taskQueue: queueNameForEnv(process.env.NODE_ENV!),
      args: [spec],
    });
    result = await workflow.result();
  } catch (e) {
    if (
      e instanceof WorkflowFailedError &&
      e.cause instanceof ApplicationFailure
    ) {
      const { type: name, cause, message } = e.cause;
      return {
        isError: true,
        message: `Hetzner connection failed: ${message}`,
      };
    }
    throw e;
  }
  redirect(`/dashboard/clusters`);
}

export async function deleteHetznerCluster(
  prevState: serverActionState,
  formData: FormData
): Promise<serverActionState> {
  const clusterId: string = formData.get("clusterId") as string;
  const cluster: HetznerCluster | undefined = await db
    .select()
    .from(hetznerClusters)
    .where(eq(hetznerClusters.id, clusterId))
    .then((rows) => rows[0] || undefined);
  if (!cluster) {
    return { isError: true, message: "Cluster not found." };
  }
  const temporalClient = await createTemporalClient;

  try {
    // cancel createHetznerCluster and provisionHetznerCluster workflows, in case either is still running
    try {
      await temporalClient.workflow
        .getHandle(`createHetznerCluster-${cluster.name}`)
        .terminate("Cluster deletion requested");
    } catch (e) {
      if (e instanceof WorkflowNotFoundError) {
        // ignore
      } else {
        throw e;
      }
    }
    try {
      await temporalClient.workflow
        .getHandle(`provisionHetznerCluster-${cluster.name}`)
        .terminate("Cluster deletion requested");
    } catch (e) {
      if (e instanceof WorkflowNotFoundError) {
        // ignore
      } else {
        throw e;
      }
    }
    const workflow = await temporalClient.workflow.start(DeleteHetznerCluster, {
      workflowId: `deleteHetznerCluster-${cluster.name}`,
      taskQueue: queueNameForEnv(process.env.NODE_ENV!),
      args: [{ clusterId }],
    });
    await workflow.result();
  } catch (e) {
    if (
      e instanceof WorkflowFailedError &&
      e.cause instanceof ApplicationFailure
    ) {
      const { type: name, cause, message } = e.cause;
      return {
        isError: true,
        message: `Creating cluster in Hetzner failed: ${message}`,
      };
    }
    throw e;
  }
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
