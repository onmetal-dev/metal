import { db } from "@/app/server/db";
import { Team, User, teams, users, usersToTeams } from "@/app/server/db/schema";
import { ClerkClient } from "@clerk/clerk-sdk-node";
import { eq } from "drizzle-orm";
import { Context, MiddlewareHandler, Next } from "hono";
import { HTTPException } from "hono/http-exception";
import z from "zod";

export function getUser(c: Context): User {
  return c.get("metalUser");
}

type ClerkAuth = ReturnType<
  Awaited<ReturnType<ClerkClient["authenticateRequest"]>>["toAuth"]
>;

export const unauthorizedResponse = {
  error: { name: "unauthorized", message: "unauthorized" },
};

const unauthorizedRes = new Response(JSON.stringify(unauthorizedResponse), {
  status: 401,
  headers: {
    "Content-Type": "application/json",
  },
});

export const userMiddleware: MiddlewareHandler = async (
  c: Context,
  next: Next
) => {
  const clerkAuth = c.get("clerkAuth") as ClerkAuth;
  if (!clerkAuth?.userId) {
    throw new HTTPException(401, { res: unauthorizedRes });
  }
  const user: User | undefined = await db.query.users.findFirst({
    where: eq(users.clerkId, clerkAuth.userId),
  });
  if (!user) {
    throw new HTTPException(401, { res: unauthorizedRes });
  }
  c.set("metalUser", user);
  await next();
};

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
