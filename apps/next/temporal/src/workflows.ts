import {
  ActivityFailure,
  ApplicationFailure,
  WorkflowInterceptorsFactory,
  proxyActivities,
} from "@temporalio/workflow";
import {
  OpenTelemetryInboundInterceptor,
  OpenTelemetryOutboundInterceptor,
} from "@temporalio/interceptors-opentelemetry/lib/workflow";
import type * as activities from "./activities"; // purely for type safety
import { HetznerProject, HetznerProjectSpec } from "@db/schema";

export const interceptors: WorkflowInterceptorsFactory = () => ({
  inbound: [new OpenTelemetryInboundInterceptor()],
  outbound: [new OpenTelemetryOutboundInterceptor()],
});

const { createHetznerProject } = proxyActivities<typeof activities>({
  startToCloseTimeout: "1 minute",
});

export async function CreateHetznerProject(
  spec: HetznerProjectSpec
): Promise<HetznerProject> {
  try {
    return await createHetznerProject(spec);
  } catch (e) {
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
