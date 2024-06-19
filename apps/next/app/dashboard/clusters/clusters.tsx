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
import { HetznerCluster } from "@/app/server/db/schema";
import { useKeyPressEvent } from "react-use";
import ReactCountryFlag from "react-country-flag";
import hetznerLocations from "@/lib/hcloud/locations";
import { forwardRef, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { useCommandItems } from "@/components/CommandMenu";
import { useRouter } from "next/navigation";
import { deleteHetznerCluster } from "./actions";
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

export function Clusters({ clusters }: ClustersProps) {
  const [focusedClusterIdx, setFocusedClusterIdx] = useState<number>(0);
  const [focusMode, setFocusMode] = useState<"mouse" | "keyboard">("mouse");
  const tableBodyRef = useRef<HTMLDivElement>(null);
  // have a ref that tracks whether cmd is pressed, so that we can distinguish cmd+k from just k
  const cmdPressed = useRef(false);
  useKeyPressEvent(
    "Meta",
    () => {
      cmdPressed.current = true;
    },
    () => {
      cmdPressed.current = false;
    }
  );
  // on j or k, switch focus mode to keyboard and set the focused cluster index
  useKeyPressEvent("j", () => {
    if (focusMode !== "keyboard") {
      setFocusMode("keyboard");
    }
    if (focusedClusterIdx + 1 < clusters.length) {
      setFocusedClusterIdx(focusedClusterIdx + 1);
    }
  });
  useKeyPressEvent("k", (event) => {
    // make sure this isn't a cmd-k
    if (cmdPressed.current) {
      return;
    }
    if (focusMode !== "keyboard") {
      setFocusMode("keyboard");
    }
    if (focusedClusterIdx > 0) {
      setFocusedClusterIdx(focusedClusterIdx - 1);
    }
  });

  const { addCommandItem, removeCommandItem, setGroupPriority } =
    useCommandItems();
  const router = useRouter();
  useEffect(() => {
    setGroupPriority("Cluster Actions", 99);
    addCommandItem({
      group: "Cluster Actions",
      label: "Create Cluster",
      onSelect: () => {
        router.push("/dashboard/clusters/new");
      },
    });
    return () => {
      removeCommandItem("Create Cluster");
    };
  }, [addCommandItem, removeCommandItem, setGroupPriority, router]);
  const deleteClusterFormRef = useRef<HTMLFormElement>(null);
  useEffect(() => {
    setGroupPriority("Selected Cluster Actions", 100);
    addCommandItem({
      group: "Selected Cluster Actions",
      label: "Delete Cluster",
      onSelect: () => {
        deleteClusterFormRef.current?.requestSubmit();
      },
    });
    return () => {
      removeCommandItem("Delete Cluster");
    };
  }, [
    addCommandItem,
    setGroupPriority,
    removeCommandItem,
    router,
    focusedClusterIdx,
  ]);

  return (
    <div className="w-full flex flex-col">
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
                    <Plus className="h-5 w-5 pr-1" />
                    <span>Create Cluster</span>
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">Create Cluster</TooltipContent>
            </Tooltip>
          </div>
        </div>

        {/* theader */}
        <div
          className={cn(
            "flex flex-row h-10 px-8 items-center bg-background/60 rounded-t-[7px] shadow-xl text-xs text-muted-foreground",
            focusedClusterIdx !== 0 && "border-b border-muted"
          )}
        >
          <div className="flex align-center w-4/12 my-auto">
            <h3>Name</h3>
          </div>
          <div className="flex align-center w-4/12 my-auto">
            <h3>Location</h3>
          </div>
          <div className="flex align-center w-4/12 my-auto">
            <h3>Status</h3>
          </div>
        </div>
        {/* tbody */}
        <div
          ref={tableBodyRef}
          className="bg-background rounded-b-[7px] shadow-2xl mb-10"
        >
          {/* trow */}
          {clusters.map((cluster, idx) => (
            <div
              key={cluster.id}
              onMouseMove={() => {
                if (focusMode !== "mouse") {
                  setFocusedClusterIdx(idx);
                  setFocusMode("mouse");
                }
              }}
              onMouseEnter={() => {
                setFocusedClusterIdx(idx);
              }}
              className={cn(
                "h-11 border-muted rounded-none",
                idx !== clusters.length - 1 && idx !== focusedClusterIdx - 1
                  ? "border-b"
                  : "",
                idx !== focusedClusterIdx ? "text-muted-foreground" : "",
                idx === focusedClusterIdx
                  ? focusMode === "mouse"
                    ? "border-2 rounded-sm border-muted-foreground/30"
                    : "border-2 rounded-sm border-primary/60"
                  : ""
              )}
              style={
                idx === focusedClusterIdx && focusMode === "keyboard"
                  ? { borderStyle: "ridge" }
                  : {}
              }
            >
              <Link
                href={`/dashboard/clusters/${cluster.id}`}
                className="text-sm px-8 flex items-center h-full"
              >
                <div className="flex align-center w-4/12">
                  <h3>{cluster.name}</h3>
                </div>
                <div className="flex align-center w-4/12">
                  <span>
                    {FlagForLocation(cluster.location)} {cluster.location}
                  </span>
                </div>
                <div className="flex align-center w-4/12">
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
                  </div>
                </div>
              </Link>
            </div>
          ))}
        </div>

        {/* 
            {selectedClusterIdx !== null && (
              <div className="mr-2">
                <DeleteClusterButton
                  ref={deleteClusterFormRef}
                  clusterId={clusters[selectedClusterIdx]!.id}
                />
              </div>
            )}
       */}
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
