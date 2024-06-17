import { db } from "@/app/server/db";
import {
  HetznerProject,
  HetznerProjectSpec,
  hetznerProjectSpec,
  hetznerProjects,
  selectHetznerProjectSchema,
} from "@/app/server/db/schema";
import { queueNameForEnv } from "@/lib/constants";
import { createTemporalClient } from "@/lib/temporal-client";
import { CreateHetznerProject } from "@/temporal/src/workflows";
import { DeleteHetznerProject } from "@/temporal/src/workflows/deleteHetznerProject";
import {
  getUser,
  idSchema,
  responseSpecs,
  unauthorizedResponse,
  userTeams,
} from "@api/shared";
import { createRoute, type OpenAPIHono } from "@hono/zod-openapi";
import { ApplicationFailure, WorkflowFailedError } from "@temporalio/client";
import { and, eq, inArray } from "drizzle-orm";
import { type Context } from "hono";
import { z } from "zod";

const paramsProjectIdSchema = z.object({
  projectId: idSchema.openapi({
    param: {
      name: "projectId",
      in: "path",
    },
    example: "3OHY5rQEfrc1vOpFrJ9q3r",
  }),
});

type ParamsProjectId = z.infer<typeof paramsProjectIdSchema>;

export default function hetznerProjectsRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "getHetznerProject",
      path: "/hetzner/projects/{projectId}",
      request: {
        params: paramsProjectIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectHetznerProjectSchema.openapi("HetznerProject"),
          "Get a Hetzner project"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
        404: responseSpecs[404],
      },
    }),
    // @ts-ignore since hono can't figure this out
    async (c: Context) => {
      const user = getUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }

      const teams = await userTeams(user.id);
      const { projectId } = (c.req.valid as (type: string) => ParamsProjectId)(
        "param"
      );
      const project: HetznerProject | undefined =
        await db.query.hetznerProjects.findFirst({
          where: and(
            eq(hetznerProjects.id, projectId),
            inArray(
              hetznerProjects.teamId,
              teams.map((t) => t.id)
            )
          ),
        });
      if (!project) {
        return c.json(
          { error: { name: "not_found", message: "Project not found" } },
          404
        );
      }
      return c.json(project);
    }
  );

  app.openapi(
    createRoute({
      method: "get",
      operationId: "getHetznerProjects",
      path: "/hetzner/projects",
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          z
            .array(selectHetznerProjectSchema.openapi("HetznerProject"))
            .openapi("HetznerProjects"),
          "Get all Hetzner projects"
        ),
        401: responseSpecs[401],
      },
    }),
    // @ts-ignore since hono can't figure this out
    async (c: Context) => {
      const user = getUser(c);
      const teams = await userTeams(user.id);
      return c.json(
        await db.query.hetznerProjects.findMany({
          where: inArray(
            hetznerProjects.teamId,
            teams.map((t) => t.id)
          ),
        })
      );
    }
  );

  app.openapi(
    createRoute({
      method: "put",
      operationId: "createHetznerProject",
      path: "/hetzner/projects/{projectId}",
      request: {
        params: paramsProjectIdSchema,
        body: {
          content: {
            "application/json": {
              schema: hetznerProjectSpec,
            },
          },
        },
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectHetznerProjectSchema.openapi("HetznerProject"),
          "Create a Hetzner project"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    // @ts-ignore since hono can't figure this out
    async (c: Context) => {
      const user = getUser(c);
      const spec: HetznerProjectSpec = (
        c.req.valid as (type: string) => HetznerProjectSpec
      )("json");

      // id has to be client-generated for idempotent PUT hetzner/projects/{id} to work
      const { projectId } = (c.req.valid as (type: string) => ParamsProjectId)(
        "param"
      );
      if (spec.id && spec.id !== projectId) {
        return c.json(
          {
            error: {
              name: "project_id_mismatch",
              message: `projectId in URL ${projectId} does not match body spec.id ${spec.id}`,
            },
          },
          400
        );
      }
      if (spec.creatorId !== user.id) {
        return c.json(
          {
            error: {
              name: "user_id_creator_id_mismatch",
              message: `userId ${user.id} does not match creatorId ${spec.creatorId}`,
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

      const temporalClient = await createTemporalClient;
      try {
        const workflow = await temporalClient.workflow.start(
          CreateHetznerProject,
          {
            workflowId: `createHetznerProject-${spec.hetznerName}`,
            taskQueue: queueNameForEnv(process.env.NODE_ENV!),
            args: [{ ...spec, id: projectId }],
          }
        );
        const result: HetznerProject = await workflow.result();
        return c.json(result);
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

  app.openapi(
    createRoute({
      method: "delete",
      operationId: "deleteHetznerProject",
      path: "/hetzner/projects/{projectId}",
      request: {
        params: paramsProjectIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](z.object({}), "Hetzner project deleted"),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    // @ts-ignore since hono can't figure this out
    async (c: Context) => {
      const user = getUser(c);
      const { projectId } = (c.req.valid as (type: string) => ParamsProjectId)(
        "param"
      );

      // pull teams for user and make sure project is part of one of their teams
      const teams = await userTeams(user.id);
      const project: HetznerProject | undefined =
        await db.query.hetznerProjects.findFirst({
          where: and(
            eq(hetznerProjects.id, projectId),
            inArray(
              hetznerProjects.teamId,
              teams.map((t) => t.id)
            )
          ),
        });
      if (!project) {
        return c.json(
          { error: { name: "not_found", message: "Project not found" } },
          404
        );
      }

      const temporalClient = await createTemporalClient;
      try {
        const workflow = await temporalClient.workflow.start(
          DeleteHetznerProject,
          {
            workflowId: `deleteHetznerProject-${projectId}`,
            taskQueue: queueNameForEnv(process.env.NODE_ENV!),
            args: [{ projectId: project.id }],
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
