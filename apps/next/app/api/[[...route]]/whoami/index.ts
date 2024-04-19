import { db } from "@/app/server/db";
import {
  users,
  teams,
  selectTeamSchema,
  selectUserSchema,
  usersToTeams,
} from "@/app/server/db/schema";
import { clerkClient } from "@clerk/nextjs";
import { decodeJwt } from "@clerk/nextjs/server";
import { eq } from "drizzle-orm";
import { type OpenAPIHono, createRoute, z } from "@hono/zod-openapi";
import { type Context } from "hono";

const whoAmISchema = z
  .object({
    token: z.string(),
    user: selectUserSchema,
    teams: z.array(selectTeamSchema),
  })
  .openapi("WhoAmI");

export default function whoami(app: OpenAPIHono) {
  app.openapi(
    createRoute({
      method: "get",
      operationId: "whoami",
      path: "/whoami",
      request: {},
      security: [{ bearerAuth: [] }],
      responses: {
        200: {
          content: {
            "application/json": {
              schema: whoAmISchema,
            },
          },
          description: "Retrieve information about the authenticated user",
        },
        401: {
          description: "Unauthorized",
          content: {
            "application/json": {
              schema: z.object({
                error: z.string(),
              }),
            },
          },
        },
      },
    }),
    // @ts-ignore because hono is bad at matching the return of this function with the schema
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
        return c.json({ error: "not authorized" }, 401);
      }
      const userTeams = await db
        .select({ team: teams })
        .from(usersToTeams)
        .where(eq(usersToTeams.userId, user.id))
        .rightJoin(teams, eq(usersToTeams.teamId, teams.id))
        .then((rows) => rows.map((row) => row.team));
      return c.json({
        token: authStatus.token,
        user,
        teams: userTeams,
      });
    }
  );
}
