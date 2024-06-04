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

type ParamsForFindCreateUserTeam = {
  userFirstName: string;
  userId: string;
  userClerkId: string;
  userClerkOrganizationId: string | null | undefined;
};
async function findCreateUserTeam({
  userFirstName,
  userId,
  userClerkId,
  userClerkOrganizationId,
}: ParamsForFindCreateUserTeam): Promise<Team> {
  /*
  - If a user is already has a Metal account and then uses Clerk's <OrganizationSwitcher />
  to create a new org, the Clerk org will exist but won't have an associated Metal team. This
  function will handle that.
  - If a user is newly registered, they won't have any organizations or teams. This
  function will handle that too.
   */
  if (userClerkOrganizationId) {
    const usersPersonalTeam: Team | undefined = await db.query.teams
      .findMany({
        where: (team, { eq, and }) =>
          and(
            eq(team.clerkId, userClerkOrganizationId),
            eq(team.creatorId, userId)
          ),
      })
      .then((rows) => rows[0] || undefined);
    if (usersPersonalTeam) {
      return usersPersonalTeam;
    }
  }

  // personal team not found, create it in Clerk
  // then tx our db: insert team, insert usersToTeams
  console.log(`Creating team for User ${userId}`);
  const clerkOrg: ClerkOrganization =
    await findCreateClerkOrganizationCreatedByUser({
      userFirstName,
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
  const { orgId } = auth();
  const userPersonalTeam = await findCreateUserTeam({
    userFirstName: user.firstName,
    userId: user.id,
    userClerkId: clerkUser.id,
    userClerkOrganizationId: orgId,
  });

  // in the future we may not do this.. but in the beginning this is where the
  // onboarding happens (i.e. creating a cluster)
  return (
    <EnsureActiveOrgSetAndRedirect
      activeOrgId={userPersonalTeam.clerkId}
      redirectTo="/dashboard/clusters"
    />
  );
}
