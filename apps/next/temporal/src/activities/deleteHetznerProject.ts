import { trace } from "@opentelemetry/api";
import {
  hetznerProjects,
  HetznerProject,
  Team,
  teams,
  hetznerClusters,
} from "@db/schema";
import { db } from "@db/index";
import { eq } from "drizzle-orm";
import { serviceName } from "@lib/constants";
import { ApplicationFailure } from "@temporalio/activity";
import createClient from "openapi-fetch";
import type { paths } from "@lib/hcloud";

export async function deleteHetznerProject({
  projectId,
}: {
  projectId: string;
}): Promise<void> {
  return await trace
    .getTracer(serviceName)
    .startActiveSpan("deleteHetznerProject", async (span) => {
      const project: HetznerProject | undefined = await db
        .select()
        .from(hetznerProjects)
        .where(eq(hetznerProjects.id, projectId))
        .then((result) => result[0] || undefined);
      if (!project) {
        throw new Error("Project not found");
      }
      const team: Team | undefined = await db
        .select()
        .from(teams)
        .where(eq(teams.id, project.teamId))
        .then((result) => result[0] || undefined);
      if (!team) {
        throw new Error("Team not found");
      }
      const clusters = await db
        .select()
        .from(hetznerClusters)
        .where(eq(hetznerClusters.hetznerProjectId, project.id))
        .then((result) => result || []);
      if (clusters.length > 0) {
        throw ApplicationFailure.create({
          type: "hetzner_project_has_clusters",
          message: `Project with id ${projectId} has clusters, delete those first`,
          nonRetryable: true,
        });
      }

      const spanAttributes = {
        hetznerProjectId: projectId,
        hetznerName: project.hetznerName,
        teamId: project.teamId,
        creatorId: project.creatorId,
      };
      span.setAttributes(spanAttributes);
      const client = createClient<paths>({
        headers: { Authorization: `Bearer ${project.hetznerApiToken}` },
        baseUrl: "https://api.hetzner.cloud/v1",
      });

      // find/delete the ssh key we created for the project
      const { data, error: getError } = await client.GET("/ssh_keys");
      if (getError) {
        throw getError;
      }
      const existingKey = data.ssh_keys.find(
        (key: any) => key.name === `metal-${projectId}`
      );
      if (existingKey) {
        await client.DELETE("/ssh_keys/{id}", {
          params: { path: { id: existingKey.id } },
        });
      }
      await db.delete(hetznerProjects).where(eq(hetznerProjects.id, projectId));
    });
}
