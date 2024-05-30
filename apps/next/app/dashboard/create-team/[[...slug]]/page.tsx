import { CreateOrganization } from "@clerk/nextjs";

export default function CreateTeamPage() {
  return <CreateOrganization afterCreateOrganizationUrl="/dashboard" />;
}
