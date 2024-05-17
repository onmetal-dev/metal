import {
  ActivityFailure,
  ApplicationFailure,
  proxyActivities,
} from "@temporalio/workflow";
import type * as activities from "../activities"; // purely for type safety

// todo: make an entitry workflow? https://community.temporal.io/t/listening-to-event-streams-in-a-workflow/10677/2
const { provisionHetznerCluster } = proxyActivities<typeof activities>({
  startToCloseTimeout: "20 minutes",
});
export async function ProvisionHetznerCluster({
  clusterId,
}: {
  clusterId: string;
}): Promise<void> {
  try {
    await provisionHetznerCluster({ clusterId });
  } catch (e) {
    if (e instanceof ActivityFailure && e.cause instanceof ApplicationFailure) {
      throw e.cause;
    }
    throw e;
  }
}
