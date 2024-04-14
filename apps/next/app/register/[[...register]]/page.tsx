import { SlimLayout } from "@/components/SlimLayout";
import { SignUp } from "@clerk/nextjs";

export default function Register() {
  return (
    <SlimLayout>
      <SignUp />
    </SlimLayout>
  );
}
