import { serviceName } from "@/lib/constants";
import { OpenAPIHono } from "@hono/zod-openapi";
import { HTTPException } from "hono/http-exception";
import { handle } from "hono/vercel";
import applicationsRoutes from "./applications";
import environmentsRoutes from "./environments";
import hetznerClustersRoutes from "./hetzner/clusters";
import hetznerProjectsRoutes from "./hetzner/projects";
import { idSchema } from "./shared";
import teamRoutes from "./teams";
import { otelTracer } from "./tracing";
import upRoutes from "./up";
import whoami from "./whoami";
export const runtime = "nodejs";

// After deliberating between trpc and hono I decided hono since it was
// much, much more straightforward for getting a basic openapi setup going
const app = new OpenAPIHono().basePath("/api");
app.use("*", otelTracer(serviceName));

// implement onError so that we pass back meaningful 500 responses
function getErrorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
app.onError((err, c) => {
  if (err instanceof HTTPException) {
    return err.getResponse();
  }
  return c.json(
    {
      error: {
        name: "internal",
        message: getErrorMessage(err),
        stack: process.env.NODE_ENV !== "production" ? err.stack : undefined,
      },
    },
    500
  );
});

whoami(app);
hetznerProjectsRoutes(app);
hetznerClustersRoutes(app);
applicationsRoutes(app);
teamRoutes(app);
environmentsRoutes(app);
upRoutes(app);

const securitySchemeKey = "bearerAuth";
app.openAPIRegistry.registerComponent("securitySchemes", securitySchemeKey, {
  type: "http",
  scheme: "bearer",
  bearerFormat: "JWT",
  description: "Bearer token",
});
app.openAPIRegistry.register("IDSchema", idSchema);

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
      url: "https://www.onmetal.dev",
      description: "Production URL",
    },
    {
      url: "http://localhost:3000",
      description: "Development URL",
    },
  ],
});

const h = handle(app);

export { h as DELETE, h as GET, h as POST, h as PUT };
