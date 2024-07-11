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

export async function mustGetClerkActiveOrg(): Promise<ClerkOrganization> {
  const clerkAuth: AuthObject = auth();
  if (!clerkAuth.orgSlug) {
    throw new Error("Clerk auth() returned null unexpectedly");
  }
  const clerkOrg: ClerkOrganization | null =
    await clerkClient().organizations.getOrganization({
      slug: clerkAuth.orgSlug,
    });
  if (!clerkOrg) {
    throw new Error("Clerk getOrganization() returned null unexpectedly");
  }
  return clerkOrg;
}

export async function findCreateClerkOrganizationCreatedByUser({
  name,
  createdByClerkId,
}: {
  name: string;
  createdByClerkId: string;
}): Promise<ClerkOrganization> {
  const { data: orgs } =
    await clerkClient().users.getOrganizationMembershipList({
      userId: createdByClerkId,
    });
  for (const org of orgs) {
    if (
      org.organization.name === name &&
      org.organization.createdBy === createdByClerkId
    ) {
      return org.organization;
    }
  }
  return await clerkClient().organizations.createOrganization({
    name,
    createdBy: createdByClerkId,
  });
}
