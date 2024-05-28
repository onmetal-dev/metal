import { SlimLayout } from "@/components/SlimLayout";
import { SignIn } from "@clerk/nextjs";

export default function Login({
  searchParams,
}: {
  searchParams: { next: string | undefined };
}) {
  return (
    <SlimLayout>
      <SignIn fallbackRedirectUrl={searchParams.next} />
    </SlimLayout>
  );
}
