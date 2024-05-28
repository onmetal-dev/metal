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

export async function findCreateClerkOrganizationCreatedByUser({
  name,
  createdByClerkId,
}: {
  name: string;
  createdByClerkId: string;
}): Promise<ClerkOrganization> {
  const { data: orgs } = await clerkClient.users.getOrganizationMembershipList({
    userId: createdByClerkId,
  });
  /* TODO-MET-30:
  const { data: orgs, totalCount } =
    await clerkClient.users.getOrganizationMembershipList({
      userId: createdByClerkId,
    });
  let hasReadAllUserOrgs = orgs.length >= totalCount;
  while (!hasReadAllUserOrgs) {
    const moreData = await clerkClient.users.getOrganizationMembershipList({
      userId: createdByClerkId,
      offset: orgs.length,
    });
    orgs.push(...moreData.data);
    hasReadAllUserOrgs = orgs.length >= totalCount;
  }
  */
  for (const org of orgs) {
    if (org.organization.name === name) {
      return org.organization;
    }
  }
  return await clerkClient.organizations.createOrganization({
    name,
    createdBy: createdByClerkId,
  });
}
