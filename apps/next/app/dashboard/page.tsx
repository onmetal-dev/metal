import { redirect } from "next/navigation";
import { auth, currentUser } from "@clerk/nextjs";
import { db } from "@/app/server/db";
import { User, UserInsert, users } from "@/app/server/db/schema";
import { eq } from "drizzle-orm";

export default async function Page() {
  // this is the post-sign-in url. The following is a hack to find/create a user object in our db
  // The Clerk-blessed way to do this is with webhooks instead
  const clerkId: string | null = auth().userId;
  if (!clerkId) {
    throw new Error(
      "No userId found... this should be an authenticated route..."
    );
  }
  const clerkUser = await currentUser();
  if (!clerkUser) {
    throw new Error(
      "No clerk user found... this should be an authenticated route..."
    );
  }
  if (clerkUser.emailAddresses.length === 0) {
    throw new Error(
      "No email address found. This should be required in how we configure Clerk"
    );
  }
  const clerkEmail = clerkUser.emailAddresses[0];
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
  await db.insert(users).values(userInsert).onConflictDoUpdate({
    target: users.id,
    set: userInsert,
  });

  const user: User | null = await db
    .select()
    .from(users)
    .where(eq(users.clerkId, clerkId))
    .limit(1)
    .then((rows) => rows[0] || null);

  // in the future we may not do this.. but in the beginning this is where the
  // onboarding happens (i.e. creating a cluster)
  redirect("/dashboard/infrastructure");
  return <div>TODO</div>;
}
