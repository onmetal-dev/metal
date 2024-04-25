import { handle } from "hono/vercel";
import { OpenAPIHono } from "@hono/zod-openapi";
import { otelTracer } from "./tracing";
import { serviceName } from "@/lib/constants";

export const config = {
  runtime: "nodejs",
};

// After deliberating between trpc and hono I decided hono since it was
// much, much more straightforward for getting a basic openapi setup going
const app = new OpenAPIHono().basePath("/api");
app.use("*", otelTracer(serviceName));

import whoami from "./whoami";
whoami(app);

const securitySchemeKey = "bearerAuth";
app.openAPIRegistry.registerComponent("securitySchemes", securitySchemeKey, {
  type: "http",
  scheme: "bearer",
  bearerFormat: "JWT",
  description: "Bearer token",
});

// The OpenAPI documentation will be available at /doc
app.doc("/doc", {
  openapi: "3.0.0",
  info: {
    version: "0.0.1",
    title: "Metal API",
    contact: {
      email: "support@onmetal.dev",
    },
  },
  security: [
    {
      [securitySchemeKey]: [],
    },
  ],
  servers: [
    {
      url:
        process.env.NODE_ENV === "production"
          ? "https://www.onmetal.dev/api"
          : "http://localhost:3000/api",
    },
  ],
});

const h = handle(app);

export { h as GET, h as POST, h as PUT, h as DELETE };