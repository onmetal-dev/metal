import { instrumentUserServerSide, mustGetUser } from "@/app/server/user";
import { fetchApplicationsBuildsDeploymentsEnvironments } from "./actions";
import Component from "./component";
import { InstrumentUserClientSide } from "@/components/Instrumentation";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default async function Page() {
  const userInstrumentation = await instrumentUserServerSide(
    await mustGetUser()
  );
  const { applications, builds, deployments, environments } =
    await fetchApplicationsBuildsDeploymentsEnvironments();
  return (
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
  );
}
