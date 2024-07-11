"use client";
import Link from "next/link";
import { Loader2, Minus, Plus } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  FocusItem,
  FocusItemCell,
  FocusItems,
  FocusList,
  FocusListHead,
  FocusListHeader,
} from "@/components/FocusList";
import { HetznerCluster } from "@/app/server/db/schema";
import ReactCountryFlag from "react-country-flag";
import hetznerLocations from "@/lib/hcloud/locations";
import { forwardRef, useCallback, useEffect, useRef, useState } from "react";
import { addCommand } from "@/providers/CommandStoreProvider";
import { useRouter } from "next/navigation";
import { deleteHetznerCluster, fetchProjectAndClusters } from "./actions";
import { useFormState, useFormStatus } from "react-dom";
import { serverActionInitialState } from "@lib/action";
import { useToast } from "@/components/ui/use-toast";

function FlagForLocation(locationName: string): React.ReactNode {
  const locationData = hetznerLocations.locations.find(
    (l) => l.name === locationName
  );
  if (locationData?.country) {
    return <ReactCountryFlag countryCode={locationData.country} />;
  }
  return undefined;
}

interface ClustersProps {
  clusters: HetznerCluster[];
}

export function Clusters({ clusters: initialClusters }: ClustersProps) {
  const [clusters, setClusters] = useState<HetznerCluster[]>(initialClusters);
  const [focusedClusterIdx, setFocusedClusterIdx] = useState(0);
  const onFocusListChange = useCallback(
    (idx: number) => {
      setFocusedClusterIdx(idx);
    },
    [setFocusedClusterIdx]
  );

  // poll cluster data every 5s--mainly to get cluster statuses updated
  useEffect(() => {
    async function fetchData() {
      const { clusters } = await fetchProjectAndClusters();
      setClusters(clusters);
    }
    fetchData();
    const interval = setInterval(fetchData, 5000); // Poll every 5000 milliseconds (5 seconds)
    return () => clearInterval(interval); // Cleanup interval on component unmount
  }, []);
  const router = useRouter();

  addCommand({
    group: "Cluster Actions",
    label: "Create Cluster",
    priority: 100,
    onSelect: () => {
      router.push("/dashboard/clusters/new");
    },
  });

  const deleteClusterFormRef = useRef<HTMLFormElement>(null);
  addCommand({
    group: "Selected Cluster Actions",
    label: "Cluster Details",
    priority: 100,
    onSelect: () => {
      router.push(`/dashboard/clusters/${clusters[focusedClusterIdx]!.name}`);
    },
  });

  // TODO: would like to make it more obvious how destructive this action is
  // addCommand({
  //   group: "Selected Cluster Actions",
  //   label: "Delete Cluster",
  //   priority: 100,
  //   onSelect: () => {
  //     deleteClusterFormRef.current?.requestSubmit();
  //   },
  // });

  return (
    <div className="flex flex-col w-full">
      <TooltipProvider>
        <div className="flex flex-row mt-3 mb-6">
          <div className="ml-auto" />
          <div>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button asChild>
                  <Link
                    href="/dashboard/clusters/new"
                    className="h-8 rounded-sm"
                  >
                    <Plus className="w-5 h-5 pr-1" />
                    <span>Create Cluster</span>
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">Create Cluster</TooltipContent>
            </Tooltip>
          </div>
        </div>

        <div>
          <FocusList
            data={clusters}
            getHref={(cluster: HetznerCluster) =>
              `/dashboard/clusters/${cluster?.name}`
            }
            defaultFocusedIdx={0}
            onFocusListChange={onFocusListChange}
          >
            <FocusListHeader>
              <FocusListHead className="w-4/12">Name</FocusListHead>
              <FocusListHead className="w-4/12">Location</FocusListHead>
              <FocusListHead className="w-4/12">Status</FocusListHead>
            </FocusListHeader>
            <FocusItems>
              {clusters.map((cluster, index) => (
                <FocusItem key={cluster.id} index={index}>
                  <FocusItemCell className="w-4/12">
                    {cluster.name}
                  </FocusItemCell>
                  <FocusItemCell className="w-4/12">
                    <span>
                      {FlagForLocation(cluster.location)} {cluster.location}
                    </span>
                  </FocusItemCell>
                  <FocusItemCell className="w-4/12">
                    <div className="flex items-center">
                      <Badge
                        className="mr-2 text-xs"
                        variant={
                          ["destroying", "destroyed"].includes(cluster.status)
                            ? "destructive"
                            : "default"
                        }
                      >
                        {cluster.status}
                      </Badge>
                      {cluster.status === "creating" ||
                        (cluster.status === "initializing" && (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ))}
                    </div>{" "}
                  </FocusItemCell>
                </FocusItem>
              ))}
            </FocusItems>
          </FocusList>
        </div>
        <div className="hidden mr-2">
          <DeleteClusterButton
            ref={deleteClusterFormRef}
            clusterId={clusters[focusedClusterIdx]!.id}
          />
        </div>
      </TooltipProvider>
    </div>
  );
}

const DeleteClusterButton = forwardRef<HTMLFormElement, { clusterId: string }>(
  ({ clusterId }: { clusterId: string }, ref) => {
    const { toast } = useToast();
    const [state, formAction] = useFormState(
      deleteHetznerCluster,
      serverActionInitialState
    );
    useEffect(() => {
      if (state.message) {
        toast({ description: state.message });
      }
    }, [state, toast]);

    return (
      <form ref={ref} action={formAction}>
        <input
          hidden={true}
          type="text"
          name="clusterId"
          value={clusterId}
          readOnly
        />
        <DeleteClusterButtonButton />
      </form>
    );
  }
);
DeleteClusterButton.displayName = "DeleteClusterButton";

const DeleteClusterButtonButton = () => {
  const status = useFormStatus();
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="submit"
          className="bg-destructive hover:bg-destructive/80"
          disabled={status.pending}
        >
          {status.pending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Minus className="h-3.5 w-3.5" />
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent side="top">Delete Cluster</TooltipContent>
    </Tooltip>
  );
};
