import {
  Organization as ClerkOrganization,
  User as ClerkUser,
} from "@clerk/nextjs/server";
import { Team, User, teams, users } from "@db/schema";
import * as Sentry from "@sentry/nextjs";
import { eq } from "drizzle-orm";
import { mustGetClerkActiveOrg, mustGetClerkUser } from "../clerk";
import { db } from "../db";

export async function mustGetUser(): Promise<User> {
  const clerkUser: ClerkUser = await mustGetClerkUser();
  const user: User | undefined = await db.query.users.findFirst({
    where: eq(users.clerkId, clerkUser.id),
  });
  if (!user) {
    throw new Error("User not found");
  }
  return user;
}

export type UserInstrumentation = {
  id: string;
  email: string;
};

// instrumentUserServerSide should be called in server components to set up instrumentation metadata for users.
// It returns the instrumentation that it set up, which can be used to pass to client components the same / consistent information.
// As of 2024-06-28 this is the recommended way to do instrumentation in nextjs / app router: https://github.com/getsentry/sentry-javascript/discussions/10019
export function instrumentUserServerSide(user: User): UserInstrumentation {
  const instrumentation = {
    id: user.id,
    email: user.email,
  };
  Sentry.setUser(instrumentation);
  return instrumentation;
}

export async function mustGetActiveTeam(): Promise<Team> {
  const clerkOrg: ClerkOrganization = await mustGetClerkActiveOrg();
  const team: Team | undefined = await db.query.teams.findFirst({
    where: eq(teams.clerkId, clerkOrg.id),
  });
  if (!team) {
    throw new Error("Team not found");
  }
  return team;
}
