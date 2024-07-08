"use client";
import {
  FocusItem,
  FocusItemCell,
  FocusItems,
  FocusList,
  FocusListHead,
  FocusListHeader,
} from "@/components/FocusList";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Application,
  Build,
  Deployment,
  Environment,
} from "@/app/server/db/schema";
import {
  Tooltip,
  TooltipProvider,
  TooltipTrigger,
  TooltipContent,
} from "@/components/ui/tooltip"; // Assuming you have a Tooltip component
import { KeySymbol } from "@/components/ui/keyboard";
import { useEffect, useState } from "react";
import { fetchDeploymentInfo } from "./actions";

export default function Component({
  applications,
  builds,
  deployments,
  environments,
}: {
  applications: Application[];
  builds: { [appId: string]: Build[] };
  deployments: { [appId: string]: Deployment[] };
  environments: Environment[];
}) {
  // if there is a production environment, default to that one
  const productionEnvironment = environments.find(
    (env) => env.name === "production"
  );
  const [selectedEnvironment, setSelectedEnvironment] = useState(
    productionEnvironment ?? environments[0]!
  );

  const [appStatuses, setAppStatuses] = useState<{ [appId: string]: string }>(
    {}
  );
  useEffect(() => {
    const fetchStatuses = async () => {
      const statuses: { [appId: string]: string } = {};

      for (const app of applications) {
        const { latestSuccessfulDeployment, recentDeployments } =
          await fetchDeploymentInfo(app.id, selectedEnvironment.id);

        if (latestSuccessfulDeployment) {
          statuses[app.id] = "running";
        } else if (recentDeployments.length > 0) {
          statuses[app.id] = recentDeployments[0]!.rolloutStatus;
        } else {
          statuses[app.id] = "no deployments";
        }
      }
      setAppStatuses(statuses);
    };
    fetchStatuses();
  }, [applications, selectedEnvironment]);

  return (
    <TooltipProvider>
      <div className="flex flex-col mt-4 text-xs">
        <div className="flex flex-row justify-end mb-4">
          <Tooltip>
            <TooltipTrigger asChild>
              <div>
                <Select defaultValue={selectedEnvironment.id}>
                  <SelectTrigger className="w-[180px]">
                    <SelectValue placeholder="Select an environment" />
                  </SelectTrigger>
                  <SelectContent>
                    {environments.map((env) => (
                      <SelectItem
                        key={env.id}
                        value={env.id}
                        onClick={() => setSelectedEnvironment(env)}
                      >
                        {env.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </TooltipTrigger>
            <TooltipContent>
              <div>
                <span className="mr-2 text-xs">Change environment</span>
                <KeySymbol disableTooltip={true} keyName="[" />
                <KeySymbol disableTooltip={true} keyName="]" />
              </div>
            </TooltipContent>
          </Tooltip>
        </div>
        <div>
          <FocusList
            data={applications}
            getHref={(application: Application) =>
              `/dashboard/applications/${application?.id}`
            }
            defaultFocusedIdx={0}
          >
            <FocusListHeader>
              <FocusListHead className="w-4/12">Application ID</FocusListHead>
              <FocusListHead className="w-4/12">Name</FocusListHead>
              <FocusListHead className="w-4/12">Status</FocusListHead>
            </FocusListHeader>
            <FocusItems>
              {applications.map((app, index) => (
                <FocusItem key={app.id} index={index}>
                  <FocusItemCell className="w-4/12">{app.id}</FocusItemCell>
                  <FocusItemCell className="w-4/12">{app.name}</FocusItemCell>
                  <FocusItemCell className="w-4/12">
                    {appStatuses[app.id] || "Loading..."}
                  </FocusItemCell>
                </FocusItem>
              ))}
            </FocusItems>
          </FocusList>
        </div>
      </div>
    </TooltipProvider>
  );
}
