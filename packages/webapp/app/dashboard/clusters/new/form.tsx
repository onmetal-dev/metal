"use client";
import { useEffect, useRef, useState } from "react";
import { Step, type StepItem, Stepper, useStepper } from "@/components/stepper";
import { createHetznerCluster } from "./actions";
import { useFormState, useFormStatus } from "react-dom";
import {
  createHetznerClusterInitialState,
  createHetznerClusterState,
} from "./shared";
import { ServerInfo } from "./shared";
import hetznerServerTypes from "@/lib/hcloud/server_types";
import hetznerLocations from "@/lib/hcloud/locations";
import hetznerPricing from "@/lib/hcloud/pricing";
import { ChooseDatacenterStep } from "./choose-datacenter";
import { ChooseServer } from "./choose-server";
import { ChooseClusterSize } from "./choose-cluster-size";
import { ReviewAndSubmit } from "./review-cost";

function prettyAmountString(amount: number, currency: string): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency,
  }).format(amount);
}

function serversForLocation(locationName: string, arch?: string): ServerInfo[] {
  const currency = hetznerPricing.pricing.currency;
  return hetznerPricing.pricing.server_types
    .filter((serverType) => {
      return (
        serverType.prices.find((price) => price.location === locationName) &&
        (arch === "arm"
          ? serverType.name.startsWith("cax")
          : !serverType.name.startsWith("cax"))
      );
    })
    .map((serverType) => {
      const price = serverType.prices.find(
        (price) => price.location === locationName
      );
      if (!price) return null;
      const specs = hetznerServerTypes.server_types.find(
        (st) => serverType.name === st.name
      );
      if (!specs) return null;
      return {
        name: serverType.name,
        priceHourly: price.price_hourly.net,
        priceMonthly: price.price_monthly.net,
        prettyPriceHourly: prettyAmountString(
          parseFloat(price.price_hourly.net),
          currency
        ),
        prettyPriceMonthly: prettyAmountString(
          parseFloat(price.price_monthly.net),
          currency
        ),
        currency,
        cores: specs.cores,
        memory: specs.memory,
        disk: specs.disk,
      };
    })
    .filter((server): server is ServerInfo => server !== null);
}

function computeMonthlyCost(
  priceMonthly: string,
  clusterSize: number,
  currency: string
): string {
  const validClusterSize = isNaN(clusterSize) ? 0 : clusterSize;
  return prettyAmountString(
    parseFloat(priceMonthly) * validClusterSize,
    currency
  );
}

export function NewClusterForm() {
  const [state, formAction] = useFormState(
    createHetznerCluster,
    createHetznerClusterInitialState
  );
  const [datacenter, setDatacenter] = useState(
    hetznerLocations.locations[0]!.name
  );
  const [clusterSize, setClusterSize] = useState(1);
  const servers = serversForLocation(datacenter, "arm");
  const [serverType, setServerType] = useState<ServerInfo>(servers[0]!);

  // get a ref to the table so we can match width dynamically for things like the inputs below
  const tableRef = useRef<HTMLDivElement>(null);
  const [tableWidth, setTableWidth] = useState(
    494 /* this is the widrth of the table on a normal laptop screen */
  );
  const cost = computeMonthlyCost(
    serverType.priceMonthly,
    clusterSize,
    serverType.currency
  );

  useEffect(() => {
    // Selecting a new datacenter might make the selected server type invalid if the
    // server type is not supported in the new datacenter. This effect checks if the
    // current serverType is still valid for the selected datacenter. If not valid,
    // reset the serverType to the default first server in the server type list.
    const serverTypeStillValid = servers.some(
      (server) => server.name === serverType.name
    );
    if (!serverTypeStillValid && servers.length > 0) {
      setServerType(servers[0]!);
    }
  }, [serverType.name, datacenter, servers]);

  // track the width of the server table so we can set the width of the inputs below it
  useEffect(() => {
    if (tableRef?.current) {
      const handleWindowResize = () =>
        setTableWidth(tableRef.current?.offsetWidth ?? 0);
      window.addEventListener("resize", handleWindowResize);
      return () => window.removeEventListener("resize", handleWindowResize);
    }
  }, [tableRef]);

  const steps: StepItem[] = [
    { label: "Choose a datacenter" },
    { label: "Choose a server type" },
    { label: "Choose a cluster size" },
    { label: "Review the cost" },
  ];

  return (
    <div className="flex w-full flex-col gap-4 text-foreground">
      <form action={formAction}>
        <input type="hidden" name="datacenter" value={datacenter} />
        <input type="hidden" name="serverType" value={serverType.name} />
        <input type="hidden" name="clusterSize" value={clusterSize} />
        <Stepper
          orientation="vertical"
          initialStep={0}
          steps={steps}
          expandVerticalSteps
        >
          <Step {...steps[0]}>
            <div className="flex mt-2 mb-4 rounded-md text-foreground">
              <ChooseDatacenterStep
                datacenter={datacenter}
                setDatacenter={setDatacenter}
              />
            </div>
          </Step>
          <Step {...steps[1]}>
            <div className="flex mt-2 mb-4 rounded-md text-foreground">
              <ChooseServer
                ref={tableRef}
                data={servers}
                serverType={serverType}
                setServerType={setServerType}
              />
            </div>
          </Step>
          <Step {...steps[2]}>
            <div
              className="flex mt-2 mb-4 rounded-md text-foreground"
              style={{ width: `${tableWidth}px` }}
            >
              <ChooseClusterSize
                clusterSize={clusterSize}
                setClusterSize={setClusterSize}
              />
            </div>
          </Step>
          <Step {...steps[3]}>
            <div
              className="flex-col mt-2 mb-4 rounded-md text-foreground"
              style={{ width: `${tableWidth}px` }}
            >
              <ReviewAndSubmit
                serverType={serverType}
                clusterSize={clusterSize}
                cost={cost}
              />
              <Result state={state} />
            </div>
          </Step>
        </Stepper>
      </form>
    </div>
  );
}

const Result = ({ state }: { state: createHetznerClusterState }) => {
  const status = useFormStatus();
  return (
    <div className="mt-5">
      {!status.pending && state.isError ? (
        <p className="text-sm text-destructive font-bold">{state.message}</p>
      ) : null}
    </div>
  );
};
