import { db } from "@/app/server/db";
import { Team, selectTeamSchema, teams } from "@/app/server/db/schema";
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

const paramsTeamIdSchema = z.object({
  teamId: idSchema.openapi({
    param: {
      name: "teamId",
      in: "path",
    },
    example: "3OHY5rQEfrc1vOpFrJ9q3r",
  }),
});
type ParamsTeamId = z.infer<typeof paramsTeamIdSchema>;

export default function teamRoutes(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "getTeam",
      path: "/teams/{teamId}",
      request: {
        params: paramsTeamIdSchema,
      },
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](selectTeamSchema.openapi("Team"), "Get a team"),
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

      const uTeams = await userTeams(user.id);
      const { teamId } = (c.req.valid as (type: string) => ParamsTeamId)(
        "param"
      );

      const team: Team | undefined = await db.query.teams.findFirst({
        where: and(
          eq(teams.id, teamId),
          inArray(
            teams.id,
            uTeams.map((t) => t.id)
          )
        ),
      });
      if (!team) {
        return c.json(
          { error: { name: "not_found", message: "Team not found" } },
          404
        );
      }
      return c.json(team);
    }
  );

  app.openapi(
    createRoute({
      method: "get",
      operationId: "getTeams",
      path: "/teams",
      security: [{ bearerAuth: [] }],
      responses: {
        200: responseSpecs[200](
          z.array(selectTeamSchema.openapi("Team")).openapi("Teams"),
          "Get all teams"
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
      return c.json(teams);
    }
  );
}
