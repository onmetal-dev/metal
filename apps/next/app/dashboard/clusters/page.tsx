"use client";
import * as React from "react";
import { Onboarding } from "./onboarding";
import { HetznerCluster, HetznerProject } from "@/app/server/db/schema";
import { Clusters } from "./clusters";
import { NoClusters } from "./no-clusters";
import { useState, useEffect } from "react";
import { fetchProjectAndClusters } from "./actions";

export default function Page() {
  const [loaded, setLoaded] = useState(false);
  const [clusters, setClusters] = useState<HetznerCluster[]>([]);
  const [hetznerProject, setHetznerProject] = useState<
    HetznerProject | undefined
  >(undefined);

  // poll cluster data every 5s--mainly to get cluster statuses updated
  useEffect(() => {
    async function fetchData() {
      const { project, clusters } = await fetchProjectAndClusters();
      setHetznerProject(project);
      setClusters(clusters);
      setLoaded(true);
    }
    fetchData();
    const interval = setInterval(fetchData, 5000); // Poll every 5000 milliseconds (5 seconds)
    return () => clearInterval(interval); // Cleanup interval on component unmount
  }, []);

  return (
    <>
      {loaded ? (
        hetznerProject ? (
          clusters.length > 0 ? (
            <Clusters clusters={clusters} />
          ) : (
            <NoClusters />
          )
        ) : (
          <Onboarding />
        )
      ) : null}
    </>
  );
}
