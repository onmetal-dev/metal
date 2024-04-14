import {
  WorkflowInterceptorsFactory,
  proxyActivities,
  sleep,
} from "@temporalio/workflow";
import {
  OpenTelemetryInboundInterceptor,
  OpenTelemetryOutboundInterceptor,
} from "@temporalio/interceptors-opentelemetry/lib/workflow";
import type * as activities from "./activities"; // purely for type safety

export const interceptors: WorkflowInterceptorsFactory = () => ({
  inbound: [new OpenTelemetryInboundInterceptor()],
  outbound: [new OpenTelemetryOutboundInterceptor()],
});

const { purchase } = proxyActivities<typeof activities>({
  startToCloseTimeout: "1 minute",
});

export async function OneClickBuy(id: string): Promise<void> {
  const result = await purchase(id); // calling the activity
  await sleep("10 seconds"); // demo use of timer
  console.log(`Activity ID: ${result} executed!`);
}
