import {
  Organization as ClerkOrganization,
  User as ClerkUser,
} from "@clerk/nextjs/server";
import { Team, User, teams, users } from "@db/schema";
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
