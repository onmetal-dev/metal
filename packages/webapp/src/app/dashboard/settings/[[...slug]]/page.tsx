import { ContentLayout } from "@/components/dashboard/ContentLayout";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import Component from "./component";

export default function TeamSettings() {
  return (
    <ContentLayout title="Clusters">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>Dashboard</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Settings</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Team</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <Component />
    </ContentLayout>
  );
}
