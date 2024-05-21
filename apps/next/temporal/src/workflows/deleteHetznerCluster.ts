import {
  ActivityFailure,
  ApplicationFailure,
  proxyActivities,
} from "@temporalio/workflow";
import type * as activities from "../activities"; // purely for type safety

const { deleteHetznerCluster } = proxyActivities<typeof activities>({
  startToCloseTimeout: "5 minutes",
});
export async function DeleteHetznerCluster({
  clusterId,
}: {
  clusterId: string;
}): Promise<void> {
  try {
    return await deleteHetznerCluster({ clusterId });
  } catch (e) {
    if (e instanceof ActivityFailure && e.cause instanceof ApplicationFailure) {
      throw e.cause;
    }
    throw e;
  }
}
