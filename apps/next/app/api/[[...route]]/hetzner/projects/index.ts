import { clerkClient } from "@clerk/nextjs/server";
import { decodeJwt } from "@clerk/nextjs/server";
import {
  HetznerProject,
  selectHetznerProjectSchema,
  teams,
  users,
  usersToTeams,
} from "@/app/server/db/schema";
import { db } from "@/app/server/db";
import { eq } from "drizzle-orm";
import { type OpenAPIHono, createRoute } from "@hono/zod-openapi";
import { z } from "zod";
import { type Context } from "hono";
import { createTemporalClient } from "@/lib/temporal-client";
import { hetznerProjectSpec, HetznerProjectSpec } from "@/app/server/db/schema";
import { CreateHetznerProject } from "@/temporal/src/workflows";
import { ApplicationFailure, WorkflowFailedError } from "@temporalio/client";
import { queueNameForEnv } from "@/lib/constants";

const paramsSchema = z.object({
  projectId: z
    .string()
    .min(22)
    .max(22)
    .refine((val) => /^[0-9a-zA-Z]{22}$/.test(val), {
      message: "projectId must be a 22 characters long base62 string",
    })
    .openapi({
      param: {
        name: "id",
        in: "path",
      },
      example: "3OHY5rQEfrc1vOpFrJ9q3r",
    }),
});

type Params = z.infer<typeof paramsSchema>;

const errorResponseSchema = z.object({
  error: z.object({
    name: z.string(),
    message: z.string(),
  }),
});

export default function hetznerProjectsRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "put",
      operationId: "createHetznerProject",
      path: "/hetzner/projects/{projectId}",
      request: {
        params: paramsSchema,
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
        200: {
          content: {
            "application/json": {
              schema: selectHetznerProjectSchema.openapi("HetznerProject"),
            },
          },
          description: "Create a Hetzner project",
        },
        400: {
          description: "Bad request",
          content: {
            "application/json": {
              schema: errorResponseSchema,
            },
          },
        },
        401: {
          description: "Unauthorized",
          content: {
            "application/json": {
              schema: errorResponseSchema,
            },
          },
        },
      },
    }),
    // @ts-ignore since hono can't figure this out
    async (c: Context) => {
      const authStatus = await clerkClient.authenticateRequest({
        request: c.req.raw,
      });
      if (!authStatus.isSignedIn) {
        return c.json({ error: "not authorized" }, 401);
      }
      const { payload: token } = decodeJwt(authStatus.token);
      const clerkUserId = token.sub;
      const user = await db
        .select()
        .from(users)
        .where(eq(users.clerkId, clerkUserId))
        .limit(1)
        .then((rows) => rows[0] || null);
      if (!user) {
        return c.json(
          { error: { name: "not_authorized", message: "not authorized" } },
          401
        );
      }
      const spec: HetznerProjectSpec = (
        c.req.valid as (type: string) => HetznerProjectSpec
      )("json");

      // id has to be client-generated for idempotent PUT hetzner/projects/{id} to work
      const { projectId } = (c.req.valid as (type: string) => Params)("param"); // ensure it's a valid base62 uuid
      if (spec.id !== projectId) {
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
      const userTeams = await db
        .select({ team: teams })
        .from(usersToTeams)
        .where(eq(usersToTeams.userId, user.id))
        .rightJoin(teams, eq(usersToTeams.teamId, teams.id))
        .then((rows) => rows.map((row) => row.team));
      const team = userTeams.find((team) => team.id === spec.teamId);
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
            workflowId: `${user.id}-${team.id}-${spec.hetznerName}`, // for idempotency adopt a unique-enough convention on the workflow id
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
}
