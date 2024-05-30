import { SlimLayout } from "@/components/SlimLayout";
import { SignUp } from "@clerk/nextjs";
import { db } from "@db/index";
import { eq } from "drizzle-orm";
import { InvitedEmail, invitedEmails } from "@/app/server/db/schema";

export default async function Register({
  searchParams,
}: {
  searchParams: { [key: string]: string | string[] | undefined };
}) {
  let inviteId = searchParams ? searchParams["invite"] : "";
  // if inviteId is an array, we only want the first one
  inviteId = Array.isArray(inviteId) ? inviteId[0] : inviteId;
  const invite: InvitedEmail | undefined = inviteId
    ? await db.query.invitedEmails.findFirst({
        where: eq(invitedEmails.id, inviteId),
      })
    : undefined;

  return (
    <SlimLayout>
      {invite ? (
        <SignUp />
      ) : (
        <div>
          <p>
            We're currently invite-only. Please add your email to the{" "}
            <a href="/" className="text-blue-500">
              waitlist
            </a>
            .
          </p>
        </div>
      )}
    </SlimLayout>
  );
}
