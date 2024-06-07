import {
  findCreateClerkOrganizationCreatedByUser,
  mustGetClerkUser,
} from "@/app/server/clerk";
import { db } from "@/app/server/db";
import {
  Team,
  User,
  UserInsert,
  teams,
  users,
  usersToTeams,
} from "@/app/server/db/schema";
import { EnsureActiveOrgSetAndRedirect } from "@/components/EnsureActiveOrgSetAndRedirect";
import {
  auth,
  type Organization as ClerkOrganization,
  type User as ClerkUser,
} from "@clerk/nextjs/server";
import { eq } from "drizzle-orm";

async function findCreateUserWithClerkId({
  clerkId,
  userInsert,
}: {
  clerkId: string;
  userInsert: UserInsert;
}): Promise<User> {
  async function findUserByClerkId(): Promise<User | null> {
    return await db
      .select()
      .from(users)
      .where(eq(users.clerkId, clerkId))
      .limit(1)
      .then((rows) => rows[0] || null);
  }
  const user: User | null = await findUserByClerkId();
  if (user) {
    return user;
  }
  await db.insert(users).values(userInsert);
  const newUser: User | null = await findUserByClerkId();
  if (!newUser) {
    throw new Error("User not found after insert");
  }
  return newUser;
}

const ensureTeamForClerkOrg = async ({
  clerkOrgId,
  teamName,
  userId,
}: {
  clerkOrgId: string;
  teamName: string;
  userId: string;
}): Promise<void> => {
  /* Fixes an issue where if a user already has a Metal account and then uses
  Clerk's <OrganizationSwitcher /> to create a new org, the new Clerk org will
  exist but won't have an associated Metal team. This function will create a
  team for the  and then add the user to it.
  */
  let team = await db.query.teams
    .findMany({
      where: (team, { eq, and }) => and(eq(team.clerkId, clerkOrgId)),
    })
    .then((rows) => rows[0] || undefined);

  if (!team) {
    await db.transaction(async (tx) => {
      const insertions = await tx
        .insert(teams)
        .values({
          clerkId: clerkOrgId,
          name: teamName,
          creatorId: userId,
        })
        .returning();
      team = insertions[0];
      if (!team) {
        tx.rollback();
        throw new Error("User team not found despite just creating it");
      }
      await tx.insert(usersToTeams).values({
        userId,
        teamId: team.id,
      });
    });
    if (!team) {
      throw new Error("unexpeced error while ensuring team existence");
    }
  }
};

async function findCreateUserTeam({
  teamName,
  userId,
  userClerkId,
}: {
  teamName: string;
  userId: string;
  userClerkId: string;
}): Promise<Team> {
  const usersPersonalTeam: Team | undefined = await db.query.teams
    .findMany({
      where: (team, { eq, and }) =>
        and(eq(team.name, teamName), eq(team.creatorId, userId)),
    })
    .then((rows) => rows[0] || undefined);
  if (usersPersonalTeam) {
    return usersPersonalTeam;
  }

  // personal team not found, create it in Clerk
  // then tx our db: insert team, insert usersToTeams
  const clerkOrg: ClerkOrganization =
    await findCreateClerkOrganizationCreatedByUser({
      name: teamName,
      createdByClerkId: userClerkId,
    });
  let team: Team | undefined;
  await db.transaction(async (tx) => {
    await tx.insert(teams).values({
      clerkId: clerkOrg.id,
      name: clerkOrg.name,
      creatorId: userId,
    });
    team = await tx.query.teams.findFirst({
      where: (teamTable, { eq }) => eq(teamTable.clerkId, clerkOrg.id),
    });
    if (!team) {
      tx.rollback();
      throw new Error("Team not found despite just creating it");
    }
    await tx.insert(usersToTeams).values({
      userId: userId,
      teamId: team.id,
    });
  });
  if (!team) {
    throw new Error("unexpeced error creating team");
  }
  return team;
}

export default async function Page() {
  // this is the post-sign-in url. The following is a hack to find/create a user object in our db
  // The Clerk-blessed way to do this is with webhooks instead
  const clerkUser: ClerkUser = await mustGetClerkUser();
  if (clerkUser.emailAddresses.length === 0) {
    throw new Error(
      "No email address found. This should be required in how we configure Clerk"
    );
  }
  const clerkEmail = clerkUser.emailAddresses[0]!;
  if (!clerkEmail.verification) {
    throw new Error(
      "No email verification status. This should be required in how we configure Clerk"
    );
  }
  const userInsert: UserInsert = {
    clerkId: clerkUser.id,
    email: clerkEmail.emailAddress,
    firstName: clerkUser.firstName || "",
    lastName: clerkUser.lastName || "",
    emailVerified: clerkEmail.verification.status === "verified",
  };
  const user: User = await findCreateUserWithClerkId({
    clerkId: clerkUser.id,
    userInsert,
  });
  const teamName = `${user.firstName}'s Projects`;
  const userPersonalTeam = await findCreateUserTeam({
    teamName,
    userId: user.id,
    userClerkId: clerkUser.id,
  });
  const { orgId } = auth();
  if (orgId) {
    await ensureTeamForClerkOrg({
      teamName,
      userId: user.id,
      clerkOrgId: orgId,
    });
  }

  // in the future we may not do this.. but in the beginning this is where the
  // onboarding happens (i.e. creating a cluster)
  return (
    <EnsureActiveOrgSetAndRedirect
      activeOrgId={userPersonalTeam.clerkId}
      redirectTo="/dashboard/clusters"
    />
  );
}
