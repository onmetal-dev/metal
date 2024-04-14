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
      const token = await getToken();
      redirect(`${next}?token=${token}`);
      return;
    } else {
      // user is not logged in--pas them to /login?redirectUrl={next}
      console.log("redirecting to", `/login?next=${encodeURIComponent(next)}`);
      redirect(`/login?next=${encodeURIComponent(next)}`);
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
