import { Client, Connection } from "@temporalio/client";
import { remember } from "@epic-web/remember";
import { Resource } from "@opentelemetry/resources";
import { SEMRESATTRS_SERVICE_NAME } from "@opentelemetry/semantic-conventions";
import { NodeSDK } from "@opentelemetry/sdk-node";
import { OpenTelemetryWorkflowClientInterceptor } from "@temporalio/interceptors-opentelemetry";
import { serviceName } from "@/lib/constants";

export const createTemporalClient = remember(
  "createTemporalClient",
  async () => {
    // otel setup taken from https://github.com/temporalio/samples-typescript/blob/main/interceptors-opentelemetry/src/client.ts
    const resource = new Resource({
      [SEMRESATTRS_SERVICE_NAME]: serviceName,
    });
    const otel = new NodeSDK({ resource });
    await otel.start();

    const crt = Buffer.from(process.env.TEMPORAL_CLIENT_CERT_DATA!, "base64");
    const key = Buffer.from(process.env.TEMPORAL_CLIENT_KEY_DATA!, "base64");

    const connection = await Connection.connect({
      address: process.env.TEMPORAL_ADDRESS!,
      tls: { clientCertPair: { crt, key } },
    });

    return new Client({
      connection,
      namespace: process.env.TEMPORAL_NAMESPACE!,
      interceptors: {
        workflow: [new OpenTelemetryWorkflowClientInterceptor()],
      },
    });
  }
);
