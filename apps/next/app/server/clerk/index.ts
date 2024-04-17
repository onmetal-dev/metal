import {
  User as ClerkUser,
  Organization as ClerkOrganization,
  AuthObject,
  SignedInAuthObject,
  clerkClient,
  auth,
  currentUser,
} from "@clerk/nextjs/server";

export async function mustGetClerkSignedInAuth(): Promise<SignedInAuthObject> {
  const clerkAuth: AuthObject = await auth();
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
  const orgs = await clerkClient.users.getOrganizationMembershipList({
    userId: createdByClerkId,
  });
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
