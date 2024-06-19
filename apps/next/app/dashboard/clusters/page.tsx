import * as React from "react";
import { Onboarding } from "./onboarding";
import { Clusters } from "./clusters";
import { NoClusters } from "./no-clusters";
import { fetchProjectAndClusters } from "./actions";

export default async function Page() {
  const { project, clusters } = await fetchProjectAndClusters();
  return (
    <>
      {project ? (
        clusters.length > 0 ? (
          <Clusters clusters={clusters} />
        ) : (
          <NoClusters />
        )
      ) : (
        <Onboarding />
      )}
    </>
  );
}
