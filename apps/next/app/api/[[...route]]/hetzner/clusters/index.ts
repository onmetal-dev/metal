import { db } from "@/app/server/db";
import {
  HetznerCluster,
  HetznerClusterInsert,
  HetznerClusterSpec,
  HetznerClusterWithNodeGroups,
  HetznerInstanceTypeEnum,
  HetznerLocationEnum,
  HetznerNetworkZoneEnum,
  HetznerNodeGroupInsert,
  HetznerProject,
  hetznerClusterSpec,
  hetznerClusters,
  hetznerNodeGroups,
  hetznerProjects,
  selectHetznerClusterWithNodeGroupsSchema,
} from "@/app/server/db/schema";
import { queueNameForEnv } from "@/lib/constants";
import { networkZoneForLocation } from "@/lib/hcloud-helpers";
import { createTemporalClient } from "@/lib/temporal-client";
import {
  DeleteHetznerCluster,
  ProvisionHetznerCluster,
} from "@/temporal/src/workflows";
import { createRoute, type OpenAPIHono } from "@hono/zod-openapi";
import {
  adjectives,
  animals,
  colors,
  uniqueNamesGenerator,
} from "@joaomoreno/unique-names-generator";
import { ApplicationFailure, WorkflowFailedError } from "@temporalio/client";
import { and, eq, inArray } from "drizzle-orm";
import { type Context } from "hono";
import uuidBase62 from "uuid-base62";
import { z } from "zod";
import {
  authenticateUser,
  idSchema,
  responseSpecs,
  unauthorizedResponse,
  userTeams,
} from "../../shared";

const paramsClusterIdSchema = z.object({
  clusterId: idSchema.openapi({
    param: {
      name: "clusterId",
      in: "path",
    },
    example: "3OHY5rQEfrc1vOpFrJ9q3r",
  }),
});

type ParamsClusterId = z.infer<typeof paramsClusterIdSchema>;

export default function hetznerClustersRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "getHetznerCluster",
      path: "/hetzner/clusters/{clusterId}",
      request: {
        params: paramsClusterIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectHetznerClusterWithNodeGroupsSchema.openapi("HetznerCluster"),
          "Get a Hetzner cluster"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
        404: responseSpecs[404],
      },
    }),
    async (c: Context) => {
      const user = await authenticateUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }

      const teams = await userTeams(user.id);
      const { clusterId } = (c.req.valid as (type: string) => ParamsClusterId)(
        "param"
      );
      const cluster: HetznerClusterWithNodeGroups | undefined =
        await db.query.hetznerClusters.findFirst({
          where: and(
            eq(hetznerClusters.id, clusterId),
            inArray(
              hetznerClusters.teamId,
              teams.map((t) => t.id)
            )
          ),
          with: {
            nodeGroups: true,
          },
        });
      if (!cluster) {
        return c.json(
          { error: { name: "not_found", message: "Cluster not found" } },
          404
        );
      }
      return c.json(cluster);
    }
  );

  app.openapi(
    createRoute({
      method: "get",
      operationId: "getHetznerClusters",
      path: "/hetzner/clusters",
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          z
            .array(
              selectHetznerClusterWithNodeGroupsSchema.openapi("HetznerCluster")
            )
            .openapi("HetznerClusters"),
          "Get all Hetzner clusters"
        ),
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }
      const teams = await userTeams(user.id);
      return c.json(
        await db.query.hetznerClusters.findMany({
          where: inArray(
            hetznerClusters.teamId,
            teams.map((t) => t.id)
          ),
          with: {
            nodeGroups: true,
          },
        })
      );
    }
  );

  app.openapi(
    createRoute({
      method: "put",
      operationId: "createHetznerCluster",
      path: "/hetzner/clusters/{clusterId}",
      request: {
        params: paramsClusterIdSchema,
        body: {
          content: {
            "application/json": {
              schema: hetznerClusterSpec,
            },
          },
        },
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectHetznerClusterWithNodeGroupsSchema.openapi("HetznerCluster"),
          "Create a Hetzner cluster"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }
      const spec: HetznerClusterSpec = (
        c.req.valid as (type: string) => HetznerClusterSpec
      )("json");

      // id has to be client-generated for idempotent PUT hetzner/projects/{id} to work
      const { clusterId } = (c.req.valid as (type: string) => ParamsClusterId)(
        "param"
      );
      if (spec.id && spec.id !== clusterId) {
        return c.json(
          {
            error: {
              name: "cluster_id_mismatch",
              message: `clusterId in URL ${clusterId} does not match body spec.id ${spec.id}`,
            },
          },
          400
        );
      }

      // pull teams for user, make sure they are part of the team in the spec
      const teams = await userTeams(user.id);
      const team = teams.find((team) => team.id === spec.teamId);
      if (!team) {
        return c.json(
          {
            error: {
              name: "not_authorized_for_team",
              message: "not authorized",
            },
          },
          401
        );
      }
      const hetznerProject: HetznerProject | undefined =
        await db.query.hetznerProjects.findFirst({
          where: eq(hetznerProjects.teamId, team.id),
        });
      if (!hetznerProject) {
        return c.json(
          {
            error: {
              name: "not_found",
              message: "Hetzner project not found for team",
            },
          },
          400
        );
      }
      const cluster: HetznerClusterInsert = {
        id: clusterId,
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
        location: spec.location as HetznerLocationEnum,
        networkZone: networkZoneForLocation(
          spec.location
        ) as HetznerNetworkZoneEnum,
      };
      const insertResult = await db
        .insert(hetznerClusters)
        .values(cluster)
        .returning({ insertedId: hetznerClusters.id });
      if (insertResult.length !== 1) {
        return c.json(
          {
            error: {
              name: "insert_failed",
              message: "Failed to insert new cluster into database.",
            },
          },
          400
        );
      }
      const nodeGroups: HetznerNodeGroupInsert[] = spec.nodeGroups.map(
        (nodeGroupSpec) => ({
          id: uuidBase62.v4(),
          clusterId,
          type: "all",
          instanceType: nodeGroupSpec.instanceType as HetznerInstanceTypeEnum,
          minNodes: nodeGroupSpec.minNodes,
          maxNodes: nodeGroupSpec.maxNodes,
        })
      );
      await db.insert(hetznerNodeGroups).values(nodeGroups);

      const temporalClient = await createTemporalClient;
      // don't await the provision workflow since this does the bulk of the work and can take very long
      temporalClient.workflow.start(ProvisionHetznerCluster, {
        workflowId: `provisionHetznerCluster-${cluster.name}`,
        taskQueue: queueNameForEnv(process.env.NODE_ENV!),
        args: [{ clusterId }],
      });
      const clusterWithNodeGroups = await db.query.hetznerClusters.findFirst({
        where: eq(hetznerClusters.id, clusterId),
        with: {
          nodeGroups: true,
        },
      });
      return c.json(clusterWithNodeGroups);
    }
  );

  app.openapi(
    createRoute({
      method: "delete",
      operationId: "deleteHetznerCluster",
      path: "/hetzner/clusters/{clusterId}",
      request: {
        params: paramsClusterIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](z.object({}), "Hetzner cluster deleted"),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }

      const { clusterId } = (c.req.valid as (type: string) => ParamsClusterId)(
        "param"
      );

      // pull teams for user and make sure project is part of one of their teams
      const teams = await userTeams(user.id);
      const cluster: HetznerCluster | undefined =
        await db.query.hetznerClusters.findFirst({
          where: and(
            eq(hetznerClusters.id, clusterId),
            inArray(
              hetznerClusters.teamId,
              teams.map((t) => t.id)
            )
          ),
        });
      if (!cluster) {
        return c.json(
          { error: { name: "not_found", message: "Cluster not found" } },
          404
        );
      }

      const temporalClient = await createTemporalClient;
      try {
        const workflow = await temporalClient.workflow.start(
          DeleteHetznerCluster,
          {
            workflowId: `deleteHetznerCluster-${clusterId}`,
            taskQueue: queueNameForEnv(process.env.NODE_ENV!),
            args: [{ clusterId: cluster.id }],
          }
        );
        await workflow.result();
        return c.json({});
      } catch (e) {
        if (
          e instanceof WorkflowFailedError &&
          e.cause instanceof ApplicationFailure
        ) {
          const { type: name, cause, message } = e.cause;
          return c.json({ error: { name, cause, message } }, 400);
        }
        throw e;
      }
    }
  );
}
