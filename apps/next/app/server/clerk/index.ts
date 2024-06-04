import {
  type AuthObject,
  type SignedInAuthObject,
} from "@clerk/backend/internal";
import {
  Organization as ClerkOrganization,
  User as ClerkUser,
  auth,
  clerkClient,
  currentUser,
} from "@clerk/nextjs/server";

export async function mustGetClerkSignedInAuth(): Promise<SignedInAuthObject> {
  const clerkAuth: AuthObject = auth();
  if (!clerkAuth.userId) {
    throw new Error("Clerk auth() returned null unexpectedly");
  }
  return clerkAuth;
}

export async function mustGetClerkUser(): Promise<ClerkUser> {
  const clerkUser: ClerkUser | null = await currentUser();
  if (!clerkUser) {
    throw new Error("Clerk currentUser() returned null unexpectedly");
  }
  return clerkUser;
}

type ParamsForFindCreateClerkOrganization = {
  userFirstName: string;
  createdByClerkId: string;
};

/**
 * This function is used to find or create a Clerk organization for the current user.
 * It first checks if the current user's organization was created by the current user.
 * If it was, it returns that organization. If not, it creates a new organization with the name `${userFirstName}'s Projects` and returns it.
 * The check `userClerkOrg?.createdBy === userId` is used to determine if the organization was created by the user.
 * This function does not check if the user is currently a member of the organization, as it is intended to find or create an organization that was created by the user, not necessarily one that they are currently a member of.
 */
export async function findCreateClerkOrganizationCreatedByUser({
  userFirstName,
  createdByClerkId,
}: ParamsForFindCreateClerkOrganization): Promise<ClerkOrganization> {
  const { orgId, userId } = auth();
  const userClerkOrg = orgId
    ? await clerkClient.organizations.getOrganization({
        organizationId: orgId,
      })
    : undefined;
  if (userClerkOrg?.createdBy === userId) {
    return userClerkOrg;
  }

  const name = `${userFirstName}'s Projects`;
  return await clerkClient.organizations.createOrganization({
    name,
    createdBy: createdByClerkId,
  });
}
