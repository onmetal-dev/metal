import { SlimLayout } from "@/components/SlimLayout";

export default async function LoginToCLISuccess() {
  return (
    <SlimLayout>
      <h1 className="font-bold text-2xl">Success!</h1>
      <p>You are logged in to the CLI. Feel free to close this window.</p>
    </SlimLayout>
  );
}
