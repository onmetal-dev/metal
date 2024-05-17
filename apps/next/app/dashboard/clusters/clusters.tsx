"use client";
import Link from "next/link";
import { Loader2, Minus, Plus } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { serverActionInitialState } from "./shared";
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
  const [focusedClusterIdx, setFocusedClusterIdx] = useState<number | null>(
    null
  );
  const incrementFocus = () => {
    if (focusedClusterIdx === null) {
      setFocusedClusterIdx(0);
    } else if (focusedClusterIdx + 1 < clusters.length) {
      setFocusedClusterIdx(focusedClusterIdx + 1);
    }
  };
  const decrementFocus = () => {
    if (focusedClusterIdx && focusedClusterIdx > 0) {
      setFocusedClusterIdx(focusedClusterIdx - 1);
    }
  };
  useKeyPressEvent("j", incrementFocus);
  useKeyPressEvent("k", decrementFocus);

  const [selectedClusterIdx, setSelectedClusterIdx] = useState<number | null>(
    null
  );
  useKeyPressEvent("x", () => {
    setSelectedClusterIdx(focusedClusterIdx);
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
    if (selectedClusterIdx === null) {
      return;
    }
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
    selectedClusterIdx,
  ]);

  return (
    <main className="grid items-start flex-1 gap-4 py-4 pl-0 pr-4 sm:py-0 md:gap-8 lg:grid-cols-3 xl:grid-cols-3">
      <div className="grid items-start gap-4 auto-rows-max md:gap-8 lg:col-span-2">
        <TooltipProvider>
          <Card>
            <CardHeader className="flex-row">
              <div className="flex-col">
                <CardTitle>Clusters</CardTitle>
                <CardDescription className="mt-4">
                  Clusters in your Hetzner account.
                </CardDescription>
              </div>
              <div className="ml-auto"></div>
              {selectedClusterIdx !== null && (
                <div className="mr-2">
                  <DeleteClusterButton
                    ref={deleteClusterFormRef}
                    clusterId={clusters[selectedClusterIdx]!.id}
                  />
                </div>
              )}
              <div>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button asChild>
                      <Link href="/dashboard/clusters/new">
                        <Plus className="h-3.5 w-3.5" />
                      </Link>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="top">Create Cluster</TooltipContent>
                </Tooltip>
              </div>
            </CardHeader>
            <CardContent>
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead></TableHead>
                      <TableHead>Name</TableHead>
                      <TableHead className="hidden sm:table-cell">
                        Location
                      </TableHead>
                      <TableHead className="hidden sm:table-cell">
                        Status
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {clusters.map((cluster, idx) => (
                      <TableRow
                        key={cluster.id}
                        className={cn(
                          "cursor-pointer",
                          focusedClusterIdx === idx && "bg-accent"
                        )}
                        onClick={() => {
                          setSelectedClusterIdx(idx);
                          setFocusedClusterIdx(idx);
                        }}
                      >
                        <TableCell>
                          <input
                            type="radio"
                            checked={selectedClusterIdx === idx}
                            onChange={() => setSelectedClusterIdx(idx)}
                          />
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{cluster.name}</div>
                        </TableCell>
                        <TableCell className="hidden sm:table-cell">
                          {FlagForLocation(cluster.location)} {cluster.location}
                        </TableCell>
                        <TableCell className="sm:table-cell">
                          <div className="flex items-center">
                            <Badge
                              className="mr-2 text-xs"
                              variant={
                                ["destroying", "destroyed"].includes(
                                  cluster.status
                                )
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
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TooltipProvider>
      </div>
      <div>
        {/* todo: right hand side content showing more details about selected cluster? */}
      </div>
    </main>
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
