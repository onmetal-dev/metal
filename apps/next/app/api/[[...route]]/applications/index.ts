import { db } from "@/app/server/db";
import {
  Application,
  ApplicationInsert,
  ApplicationSpec,
  applicationSpec,
  applications,
  selectApplicationSchema,
} from "@/app/server/db/schema";
import {
  authenticateUser,
  idSchema,
  responseSpecs,
  unauthorizedResponse,
  userTeams,
} from "@api/shared";
import { createRoute, type OpenAPIHono } from "@hono/zod-openapi";
import { and, eq, inArray } from "drizzle-orm";
import { type Context } from "hono";
import { z } from "zod";

const paramsApplicationIdSchema = z.object({
  applicationId: idSchema.openapi({
    param: {
      name: "applicationId",
      in: "path",
    },
    example: "3OHY5rQEfrc1vOpFrJ9q3r",
  }),
});
type ParamsApplicationId = z.infer<typeof paramsApplicationIdSchema>;

export default function applicationsRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "getApplication",
      path: "/applications/{applicationId}",
      request: {
        params: paramsApplicationIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectApplicationSchema.openapi("Application"),
          "Get an application"
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
      const { applicationId } = (
        c.req.valid as (type: string) => ParamsApplicationId
      )("param");
      // const { teamId } = (c.req.valid as (type: string) => ParamsTeamId)(
      //   "query"
      // );

      const application: Application | undefined =
        await db.query.applications.findFirst({
          where: and(
            eq(applications.id, applicationId),
            inArray(
              applications.teamId,
              teams.map((t) => t.id)
            )
          ),
        });
      if (!application) {
        return c.json(
          { error: { name: "not_found", message: "Application not found" } },
          404
        );
      }
      return c.json(application);
    }
  );

  app.openapi(
    createRoute({
      method: "get",
      operationId: "getApplications",
      path: "/applications",
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          z
            .array(selectApplicationSchema.openapi("Application"))
            .openapi("Applications"),
          "Get all applications"
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
        await db.query.applications.findMany({
          where: inArray(
            applications.teamId,
            teams.map((t) => t.id)
          ),
        })
      );
    }
  );

  app.openapi(
    createRoute({
      method: "put",
      operationId: "createApplication",
      path: "/applications/{applicationId}",
      request: {
        params: paramsApplicationIdSchema,
        body: {
          content: {
            "application/json": {
              schema: applicationSpec,
            },
          },
        },
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          selectApplicationSchema.openapi("Application"),
          "Create an application"
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
      const spec: ApplicationSpec = (
        c.req.valid as (type: string) => ApplicationSpec
      )("json");

      // id client-generated for idempotent PUT
      const { applicationId } = (
        c.req.valid as (type: string) => ParamsApplicationId
      )("param");
      if (spec.id && spec.id !== applicationId) {
        return c.json(
          {
            error: {
              name: "application_id_mismatch",
              message: `applicationId in URL ${applicationId} does not match body spec.id ${spec.id}`,
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

      const application: ApplicationInsert = {
        id: applicationId,
        creatorId: user.id,
        teamId: team.id,
        name: spec.name,
      };
      const insertResult = await db
        .insert(applications)
        .values(application)
        .returning({ insertedId: applications.id });
      if (insertResult.length !== 1) {
        return c.json(
          {
            error: {
              name: "insert_failed",
              message: "Failed to insert new application into database.",
            },
          },
          400
        );
      }
      const insertedApp = await db.query.applications.findFirst({
        where: eq(applications.id, insertResult[0]!.insertedId),
      });
      return c.json(insertedApp);
    }
  );

  app.openapi(
    createRoute({
      method: "delete",
      operationId: "deleteApplication",
      path: "/applications/{applicationId}",
      request: {
        params: paramsApplicationIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](z.object({}), "Application deleted"),
        400: responseSpecs[400],
        401: responseSpecs[401],
      },
    }),
    async (c: Context) => {
      const user = await authenticateUser(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }

      const { applicationId } = (
        c.req.valid as (type: string) => ParamsApplicationId
      )("param");

      // pull teams for user and make sure application is part of one of their teams
      const teams = await userTeams(user.id);
      const application: Application | undefined =
        await db.query.hetznerClusters.findFirst({
          where: and(
            eq(applications.id, applicationId),
            inArray(
              applications.teamId,
              teams.map((t) => t.id)
            )
          ),
        });
      if (!application) {
        return c.json(
          { error: { name: "not_found", message: "Application not found" } },
          404
        );
      }
      await db.delete(applications).where(eq(applications.id, applicationId));
      return c.json({});

      // TODO: deleting an application is a bit more involved than just removing it from a db
      // have to delete all associated resources, including deployments
      //   const temporalClient = await createTemporalClient;
      //   try {
      //     const workflow = await temporalClient.workflow.start(
      //       DeleteApplication,
      //       {
      //         workflowId: `deleteApplication-${applicationId}`,
      //         taskQueue: queueNameForEnv(process.env.NODE_ENV!),
      //         args: [{ applicationId: application.id }],
      //       }
      //     );
      //     await workflow.result();
      //     return c.json({});
      //   } catch (e) {
      //     if (
      //       e instanceof WorkflowFailedError &&
      //       e.cause instanceof ApplicationFailure
      //     ) {
      //       const { type: name, cause, message } = e.cause;
      //       return c.json({ error: { name, cause, message } }, 400);
      //     }
      //     throw e;
      //   }
    }
  );
}
