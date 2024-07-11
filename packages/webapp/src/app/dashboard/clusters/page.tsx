import * as React from "react";
import { Onboarding } from "./onboarding";
import { Clusters } from "./clusters";
import { NoClusters } from "./no-clusters";
import { fetchProjectAndClusters } from "./actions";
import { instrumentUserServerSide, mustGetUser } from "@/app/server/user";
import { InstrumentUserClientSide } from "@/components/Instrumentation";
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
  const { project, clusters } = await fetchProjectAndClusters();
  return (
    <ContentLayout title="Clusters">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>Dashboard</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Clusters</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <InstrumentUserClientSide user={userInstrumentation}>
        {project ? (
          clusters.length > 0 ? (
            <Clusters clusters={clusters} />
          ) : (
            <NoClusters />
          )
        ) : (
          <Onboarding />
        )}
      </InstrumentUserClientSide>
    </ContentLayout>
  );
}
