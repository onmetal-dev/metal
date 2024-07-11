import { instrumentUserServerSide, mustGetUser } from "@/app/server/user";
import { NewClusterForm } from "./form";
import Link from "next/link";

import { ContentLayout } from "@/components/dashboard/ContentLayout";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { InstrumentUserClientSide } from "@/components/Instrumentation";

export default async function NewClusterPage() {
  const userInstrumentation = await instrumentUserServerSide(
    await mustGetUser()
  );

  return (
    <ContentLayout title="Create a Cluster">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>Dashboard</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link href="/dashboard/clusters">Clusters</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>New</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <InstrumentUserClientSide user={userInstrumentation}>
        <div className="grid flex-1 items-start gap-4 pl-0 pr-4 py-4 sm:py-0 md:gap-8 lg:grid-cols-3 mt-4">
          <div className="grid auto-rows-max items-start gap-4 lg:gap-8 lg:col-span-2">
            <div className="grid gap-4">
              <NewClusterForm />
            </div>
          </div>
          <div></div>
        </div>
      </InstrumentUserClientSide>
    </ContentLayout>
  );
}
