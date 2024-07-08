// This file configures the initialization of Sentry on the server.
// The config you add here will be used whenever the server handles a request.
// https://docs.sentry.io/platforms/javascript/guides/nextjs/
// modified in order to also send spans to local OTLP exporter https://docs.sentry.io/platforms/javascript/guides/node/tracing/instrumentation/opentelemetry/
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { Resource } from "@opentelemetry/resources";
import {
  BatchSpanProcessor,
  NodeTracerProvider,
} from "@opentelemetry/sdk-trace-node";
import * as Sentry from "@sentry/node";
import {
  SentryPropagator,
  SentrySampler,
  SentrySpanProcessor,
} from "@sentry/opentelemetry";

const sentryClient = Sentry.init({
  enabled: process.env.NODE_ENV === "production",
  dsn: "https://72c022fba7a2a5476ad24a977157f49a@o4507506600181760.ingest.us.sentry.io/4507506602999808",
  skipOpenTelemetrySetup: true,

  // Adjust this value in production, or use tracesSampler for greater control
  tracesSampleRate: 1,

  // Setting this option to true will print useful information to the console while you're setting up Sentry.
  debug: false,

  // Uncomment the line below to enable Spotlight (https://spotlightjs.com)
  // spotlight: process.env.NODE_ENV === 'development',
});

// Note: This could be BasicTracerProvider or any other provider depending on
// how you are using the OpenTelemetry SDK
const provider = new NodeTracerProvider({
  // We need our sampler to ensure the correct subset of traces is sent to Sentry
  sampler: sentryClient ? new SentrySampler(sentryClient) : undefined,
  resource: new Resource({
    "service.name": "webapp",
  }),
});

// We need a custom span processor
provider.addSpanProcessor(new SentrySpanProcessor());

// Additionally send spans to local OTLP
provider.addSpanProcessor(new BatchSpanProcessor(new OTLPTraceExporter()));

// We need a custom propagator and context manager
provider.register({
  propagator: new SentryPropagator(),
  contextManager: new Sentry.SentryContextManager(),
});

// Validate that the setup is correct
Sentry.validateOpenTelemetrySetup();
