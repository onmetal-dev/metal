import {
  OpenTelemetryInboundInterceptor,
  OpenTelemetryOutboundInterceptor,
} from "@temporalio/interceptors-opentelemetry/lib/workflow";
import { WorkflowInterceptorsFactory } from "@temporalio/workflow";

export const interceptors: WorkflowInterceptorsFactory = () => ({
  inbound: [new OpenTelemetryInboundInterceptor()],
  outbound: [new OpenTelemetryOutboundInterceptor()],
});

export * from "./workflows/createHetznerProject";
export * from "./workflows/deleteHetznerCluster";
export * from "./workflows/deleteHetznerProject";
export * from "./workflows/provisionHetznerCluster";
