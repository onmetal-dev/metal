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
  organizationId: string | null | undefined;
  userFirstName: string;
  createdByClerkId: string;
};
export async function findCreateClerkOrganizationCreatedByUser({
  organizationId,
  userFirstName,
  createdByClerkId,
}: ParamsForFindCreateClerkOrganization): Promise<ClerkOrganization> {
  const userClerkOrg = organizationId
    ? await clerkClient.organizations.getOrganization({
        organizationId,
      })
    : undefined;
  if (userClerkOrg) {
    return userClerkOrg;
  }

  const { totalCount } = await clerkClient.users.getOrganizationMembershipList({
    userId: createdByClerkId,
  });
  const newTotalCount = 1 + totalCount;
  const name = `[${newTotalCount}] ${userFirstName}'s Projects`;
  return await clerkClient.organizations.createOrganization({
    name,
    createdBy: createdByClerkId,
  });
}
