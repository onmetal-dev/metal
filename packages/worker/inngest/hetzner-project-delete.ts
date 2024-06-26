import type { paths } from "@metal/hcloud";
import { db } from "@metal/webapp/app/server/db";
import type { HetznerProject, Team } from "@metal/webapp/app/server/db/schema";
import {
  hetznerClusters,
  hetznerProjects,
  teams,
} from "@metal/webapp/app/server/db/schema";
import { serviceName } from "@metal/webapp/lib/constants";
import { trace } from "@opentelemetry/api";
import { eq } from "drizzle-orm";
import createClient from "openapi-fetch";
import { inngest } from "./client";

export const hetznerProjectDelete = inngest.createFunction(
  { id: "hetzner-project-delete" },
  { event: "hetzner-project/delete" },
  async ({ event, step }) => {
    return await trace
      .getTracer(serviceName)
      .startActiveSpan("deleteHetznerProject", async (span) => {
        const { projectId } = event.data;
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
          return {
            type: "hetzner_project_has_clusters",
            message: `Project with id ${projectId} has clusters, delete those first`,
          };
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
        await db
          .delete(hetznerProjects)
          .where(eq(hetznerProjects.id, projectId));
      });
  }
);
