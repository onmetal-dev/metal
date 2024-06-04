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
