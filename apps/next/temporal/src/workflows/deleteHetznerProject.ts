import {
  ActivityFailure,
  ApplicationFailure,
  proxyActivities,
} from "@temporalio/workflow";
import type * as activities from "../activities"; // purely for type safety

const { deleteHetznerProject } = proxyActivities<typeof activities>({
  startToCloseTimeout: "1 minute",
});

export async function DeleteHetznerProject({
  projectId,
}: {
  projectId: string;
}): Promise<void> {
  try {
    return await deleteHetznerProject({ projectId });
  } catch (e: any) {
    // if root cause is an application failure, we want to unwrap it to
    // a) make it easier to extract the user-facing type and message from the ApplicationFailure
    // b) bubble up the non-retryable flag at the Workflow level (i.e. don't retry the workflow if the activity throws a non-retryable ApplicationError)
    // this is the less-fancy version of https://www.flightcontrol.dev/blog/temporal-error-handling-in-practice
    if (e instanceof ActivityFailure && e.cause instanceof ApplicationFailure) {
      throw e.cause;
    }
    throw e;
  }
}
