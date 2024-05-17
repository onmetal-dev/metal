import { WorkflowInterceptorsFactory } from "@temporalio/workflow";
import {
  OpenTelemetryInboundInterceptor,
  OpenTelemetryOutboundInterceptor,
} from "@temporalio/interceptors-opentelemetry/lib/workflow";

export const interceptors: WorkflowInterceptorsFactory = () => ({
  inbound: [new OpenTelemetryInboundInterceptor()],
  outbound: [new OpenTelemetryOutboundInterceptor()],
});

export * from "./workflows/createHetznerProject";
export * from "./workflows/deleteHetznerProject";
export * from "./workflows/provisionHetznerCluster";
export * from "./workflows/deleteHetznerCluster";
