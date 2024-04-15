import { SlimLayout } from "@/components/SlimLayout";
import { auth } from "@clerk/nextjs";
import { redirect } from "next/navigation";

export default async function LoginToCLI({
  searchParams,
}: {
  searchParams: { next: string; success: string };
}) {
  const { userId, getToken } = auth();
  const { next } = searchParams;
  if (next) {
    if (userId) {
      // user is logged in--pass a token to the CLI
      const token = await getToken({ template: "cli" });
      redirect(`${next}?token=${token}`);
      return;
    } else {
      // we need the redirect to come back to /login-to-cli with the same next param and trigger the getToken logic
      const redirectUrl = `/login?next=${encodeURIComponent(
        `/login-to-cli?next=${encodeURIComponent(next)}`
      )}`;
      redirect(redirectUrl);
      return;
    }
  }

  // at this point there is no next, so assume this is a succesful login
  return (
    <SlimLayout>
      <h1 className="font-bold text-2xl">Success!</h1>
      <p>You are logged in to the CLI. Feel free to close this window.</p>
    </SlimLayout>
  );
}
