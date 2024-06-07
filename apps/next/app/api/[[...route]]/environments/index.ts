import { db } from "@/app/server/db";
import {
  Environment,
  EnvironmentInsert,
  EnvironmentSpec,
  environmentSpec,
  environments,
  selectEnvironmentSchema,
} from "@/app/server/db/schema";
import {
  authenticateRequest,
  idSchema,
  responseSpecs,
  unauthorizedResponse,
  userTeams,
} from "@api/shared";
import { createRoute, type OpenAPIHono } from "@hono/zod-openapi";
import { and, eq, inArray } from "drizzle-orm";
import { type Context } from "hono";
import { z } from "zod";

const paramsEnvironmentIdSchema = z.object({
  environmentId: idSchema.openapi({
    param: {
      name: "environmentId",
      in: "path",
    },
    example: "3OHY5rQEfrc1vOpFrJ9q3r",
  }),
});
type ParamsEnvironmentId = z.infer<typeof paramsEnvironmentIdSchema>;

const paramsTeamInQuerySchema = z.object({
  teamId: idSchema.nullable().openapi({
    param: {
      name: "teamId",
      in: "query",
      description: "The ID of the team to get the environments for",
    },
  }),
});
type ParamsTeamInQuery = z.infer<typeof paramsTeamInQuerySchema>;

export default function environmentRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "getEnvironment",
      path: "/environments/{environmentId}",
      request: {
        params: paramsEnvironmentIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectEnvironmentSchema.openapi("Environment"),
          "Get an environment"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
        404: responseSpecs[404],
      },
    }),
    async (c: Context) => {
      const user = await authenticateRequest(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }

      const uTeams = await userTeams(user.id);
      const { environmentId } = (
        c.req.valid as (type: string) => ParamsEnvironmentId
      )("param");

      const environment: Environment | undefined =
        await db.query.environments.findFirst({
          where: and(
            eq(environments.id, environmentId),
            inArray(
              environments.teamId,
              uTeams.map((t) => t.id)
            )
          ),
        });
      if (!environment) {
        return c.json(
          { error: { name: "not_found", message: "Environment not found" } },
          404
        );
      }
      return c.json(environment);
    }
  );

  app.openapi(
    createRoute({
      method: "get",
      operationId: "getEnvironments",
      path: "/environments",
      request: {
        query: paramsTeamInQuerySchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          z
            .array(selectEnvironmentSchema.openapi("Environment"))
            .openapi("Environments"),
          "Get all environments"
        ),
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateRequest(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }
      const uTeams = await userTeams(user.id);
      let where = [
        inArray(
          environments.teamId,
          uTeams.map((t) => t.id)
        ),
      ];
      const { teamId } = (c.req.valid as (type: string) => ParamsTeamInQuery)(
        "query"
      );
      if (teamId) {
        where.push(eq(environments.teamId, teamId));
      }
      const envs: Environment[] = await db.query.environments.findMany({
        where: and(...where),
      });
      return c.json(envs);
    }
  );

  app.openapi(
    createRoute({
      method: "put",
      operationId: "createEnvironment",
      path: "/environments/{environmentId}",
      request: {
        params: paramsEnvironmentIdSchema,
        body: {
          content: {
            "application/json": {
              schema: environmentSpec,
            },
          },
        },
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectEnvironmentSchema.openapi("Environment"),
          "Create an environment"
        ),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateRequest(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }
      const spec: EnvironmentSpec = (
        c.req.valid as (type: string) => EnvironmentSpec
      )("json");

      const { environmentId } = (
        c.req.valid as (type: string) => ParamsEnvironmentId
      )("param");
      if (spec.id && spec.id !== environmentId) {
        return c.json(
          {
            error: {
              name: "environment_id_mismatch",
              message: `environmentId in URL ${environmentId} does not match body ${spec.id}`,
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

      const environment: EnvironmentInsert = {
        id: environmentId,
        teamId: team.id,
        name: spec.name,
      };
      const insertResult = await db
        .insert(environments)
        .values(environment)
        .returning({ insertedId: environments.id });
      if (insertResult.length !== 1) {
        return c.json(
          {
            error: {
              name: "insert_failed",
              message: "Failed to insert new environment into database.",
            },
          },
          400
        );
      }
      const e = await db.query.environments.findFirst({
        where: eq(environments.id, insertResult[0]!.insertedId),
      });
      return c.json(e);
    }
  );
}
