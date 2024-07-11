import { instrumentUserServerSide, mustGetUser } from "@/app/server/user";
import { fetchApplicationsBuildsDeploymentsEnvironments } from "./actions";
import Component from "./component";
import { InstrumentUserClientSide } from "@/components/Instrumentation";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ContentLayout } from "@/components/dashboard/ContentLayout";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export default async function Page() {
  const userInstrumentation = await instrumentUserServerSide(
    await mustGetUser()
  );
  const { applications, builds, deployments, environments } =
    await fetchApplicationsBuildsDeploymentsEnvironments();
  return (
    <ContentLayout title="Applications">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>Dashboard</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Applications</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <InstrumentUserClientSide user={userInstrumentation}>
        {applications.length === 0 && (
          <Card>
            <CardHeader className="flex flex-col items-center justify-center">
              <CardTitle>No applications found!</CardTitle>
              <CardDescription>
                Deploy one with <code>metal up</code>
              </CardDescription>
            </CardHeader>
          </Card>
        )}
        {applications.length > 0 && (
          <Component
            applications={applications}
            builds={builds}
            deployments={deployments}
            environments={environments}
          />
        )}
      </InstrumentUserClientSide>
    </ContentLayout>
  );
}
