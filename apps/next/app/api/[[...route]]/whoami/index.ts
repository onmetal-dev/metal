import {
  selectTeamSchema,
  selectUserSchema,
  User,
} from "@/app/server/db/schema";
import { clerkClient } from "@clerk/nextjs";
import { type OpenAPIHono, createRoute } from "@hono/zod-openapi";
import { z } from "zod";
import { type Context } from "hono";
import {
  authenticateRequest,
  unauthorizedResponse,
  userTeams,
} from "../shared";

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
        return c.json(unauthorizedResponse, 401);
      }
      const user: User | undefined = await authenticateRequest(c);
      if (!user) {
        return c.json(unauthorizedResponse, 401);
      }
      const teams = await userTeams(user.id);
      return c.json({
        token: authStatus.token,
        user,
        teams,
      });
    }
  );
}
