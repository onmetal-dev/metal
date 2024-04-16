import { redirect } from "next/navigation";

export default function Page() {
  // in the future we may not do this.. but in the beginning this is where the
  // onboarding happens (i.e. creating a cluster)
  redirect("/dashboard/infrastructure");
  return <div>TODO</div>;
}
