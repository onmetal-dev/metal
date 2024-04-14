import { Worker, NativeConnection } from "@temporalio/worker";
import * as activities from "./activities";
import fs from "fs";
import { Resource } from "@opentelemetry/resources";
import { SEMRESATTRS_SERVICE_NAME } from "@opentelemetry/semantic-conventions";
import { NodeSDK } from "@opentelemetry/sdk-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-proto";
import {
  OpenTelemetryActivityInboundInterceptor,
  makeWorkflowExporter,
} from "@temporalio/interceptors-opentelemetry/lib/worker";

run().catch((err) => console.log(err));

console.log("DEBUG NODE ENV", process.env.NODE_ENV);

async function run() {
  // otel setup taken from https://github.com/temporalio/samples-typescript/blob/main/interceptors-opentelemetry/src/worker.ts
  const resource = new Resource({
    [SEMRESATTRS_SERVICE_NAME]:
      `metal-next` + (process.env.NODE_ENV === "development" ? "-dev" : ""),
  });
  const otel = new NodeSDK({ resource });
  await otel.start();

  const crt = Buffer.from(process.env.TEMPORAL_CLIENT_CERT_DATA!, "base64");
  const key = Buffer.from(process.env.TEMPORAL_CLIENT_KEY_DATA!, "base64");

  const connection = await NativeConnection.connect({
    address: "metal-dev.hsfvl.tmprl.cloud:7233",
    tls: {
      clientCertPair: {
        crt,
        key,
      },
    },
  });
  const exporter = new OTLPTraceExporter({
    // todo: in prod probably want this
    //url: "otel-collector.opentelemetry.svc.cluster.local:4318",
  });
  const worker = await Worker.create({
    connection,
    sinks: {
      exporter: makeWorkflowExporter(exporter, resource),
    },
    // register opentelemetry interceptors for Workflow and Activity calls
    interceptors: {
      workflowModules: [require.resolve("./workflows")],
      activityInbound: [
        (ctx) => new OpenTelemetryActivityInboundInterceptor(ctx),
      ],
    },
    namespace: "metal-dev.hsfvl",
    workflowsPath: require.resolve("./workflows"), // passed to Webpack for bundling
    activities, // directly imported in Node.js
    taskQueue: "tutorial",
  });
  try {
    await worker.run();
  } finally {
    await connection.close();
    await otel.shutdown();
  }
}
