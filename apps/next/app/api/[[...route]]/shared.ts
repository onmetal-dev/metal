import { db } from "@/app/server/db";
import { Team, User, teams, users, usersToTeams } from "@/app/server/db/schema";
import { decodeJwt } from "@clerk/backend/jwt";
import { clerkClient } from "@clerk/clerk-sdk-node";
import { eq } from "drizzle-orm";
import { Context } from "hono";
import z from "zod";

export async function authenticateUser(c: Context): Promise<User | undefined> {
  const authStatus = await clerkClient.authenticateRequest(c.req.raw);
  if (!authStatus.isSignedIn) {
    return undefined;
  }
  const { payload: token } = decodeJwt(authStatus.token);
  const clerkUserId = token.sub;
  const user: User | undefined = await db.query.users.findFirst({
    where: eq(users.clerkId, clerkUserId),
  });
  return user;
}

export async function userTeams(userId: string): Promise<Team[]> {
  return await db
    .select({ team: teams })
    .from(usersToTeams)
    .where(eq(usersToTeams.userId, userId))
    .rightJoin(teams, eq(usersToTeams.teamId, teams.id))
    .then((rows) => rows.map((row) => row.team));
}

export const idSchema = z
  .string()
  .min(22)
  .max(22)
  .refine((val) => /^[0-9a-zA-Z]{22}$/.test(val), {
    message: "projectId must be a 22 characters long base62 string",
  });

export const nameSchema = z
  .string()
  .refine((val) => /^[0-9a-zA-Z-_]+$/.test(val), {
    message:
      "name must be a string of alphanumeric characters, hyphens, or underscores",
  });

export const errorResponseSchema = z.object({
  error: z.object({
    name: z.string(),
    message: z.string(),
    issues: z.array(z.record(z.string(), z.any())),
  }),
});

export const unauthorizedResponse = {
  error: { name: "unauthorized", message: "unauthorized" },
};

export const responseSpecs = {
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
  404: {
    description: "Not found",
    content: {
      "application/json": {
        schema: errorResponseSchema,
      },
    },
  },
  200: (schema: any, description: string) => ({
    description,
    content: {
      "application/json": {
        schema,
      },
    },
  }),
};
