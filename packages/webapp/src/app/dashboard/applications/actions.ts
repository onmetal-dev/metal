"use server";
import { db } from "@/app/server/db";
import {
  Application,
  Build,
  Deployment,
  Environment,
  Team,
  applications,
  builds,
  deployments,
  environments,
  teams,
} from "@/app/server/db/schema";
import { auth } from "@clerk/nextjs/server";
import { and, desc, eq } from "drizzle-orm";

export async function fetchApplicationsBuildsDeploymentsEnvironments(): Promise<{
  applications: Application[];
  builds: { [appId: string]: Build[] };
  deployments: { [appId: string]: Deployment[] };
  environments: Environment[];
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
  const apps: Application[] = await db.query.applications.findMany({
    where: eq(applications.teamId, team.id),
  });

  const envs: Environment[] = await db.query.environments.findMany({
    where: eq(environments.teamId, team.id),
  });

  // for each application fetch up to 5 builds and 5 deployments
  let bs: { [appId: string]: Build[] } = {};
  let ds: { [appId: string]: Deployment[] } = {};
  await Promise.all(
    apps.map(async (app) => {
      const [appBuilds, appDeployments] = await Promise.all([
        db.query.builds.findMany({
          where: eq(builds.applicationId, app.id),
          orderBy: desc(builds.createdAt),
          limit: 5,
        }),
        db.query.deployments.findMany({
          where: eq(deployments.applicationId, app.id),
          orderBy: desc(deployments.createdAt),
          limit: 5,
        }),
      ]);
      bs[app.id] = appBuilds;
      ds[app.id] = appDeployments;
    })
  );
  return {
    applications: apps,
    environments: envs,
    builds: bs,
    deployments: ds,
  };
}

export async function fetchDeploymentInfo(
  applicationId: string,
  environmentId: string
): Promise<{
  latestSuccessfulDeployment: Deployment | null;
  recentDeployments: Deployment[];
}> {
  const latestSuccessfulDeployment = await db.query.deployments.findFirst({
    where: and(
      eq(deployments.applicationId, applicationId),
      eq(deployments.environmentId, environmentId),
      eq(deployments.rolloutStatus, "running")
    ),
    orderBy: desc(deployments.createdAt),
  });

  const recentDeployments = await db.query.deployments.findMany({
    where: and(
      eq(deployments.applicationId, applicationId),
      eq(deployments.environmentId, environmentId)
    ),
    orderBy: desc(deployments.createdAt),
    limit: 5,
  });

  return {
    latestSuccessfulDeployment: latestSuccessfulDeployment || null,
    recentDeployments,
  };
}
