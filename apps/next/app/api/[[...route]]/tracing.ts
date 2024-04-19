// this file borrows heavily from https://github.com/watany-dev/middleware/tree/otel/packages/open-telemetry/src/trace
// via https://github.com/honojs/hono/issues/1864
import { Span, trace } from "@opentelemetry/api";
import { Context } from "hono";
import { createMiddleware } from "hono/factory";

const recordError = (span: Span, error: unknown) => {
  if (error instanceof Error) {
    span.recordException({
      name: error.name,
      message: error.message,
      stack: error.stack,
    });
    span.setStatus({ code: 2, message: error.message });
  } else {
    const errorMessage = String(error);
    span.recordException({ message: errorMessage });
    span.setStatus({ code: 2, message: errorMessage });
  }
};

const spanContextKey = "span";
export const getSpan = (c: Context): Span => {
  return c.get(spanContextKey);
};

export const otelTracer = (
  tracerName: string,
  customAttributes?: (context: Context) => Record<string, unknown>
) => {
  return createMiddleware(async (c: Context, next: () => Promise<void>) => {
    const tracer = trace.getTracer(tracerName);
    const span = tracer.startSpan("http-request", {
      attributes: {
        "http.method": c.req.method,
        "http.url": c.req.url,
        ...customAttributes,
      },
    });
    const startTime = Date.now();
    c.set(spanContextKey, span);

    try {
      await next();
      span.setAttribute("http.status_code", c.res.status);
    } catch (error) {
      recordError(span, error);
      if (c.req.header("User-Agent")) {
        const userAgent = c.req.header("User-Agent") || "Unknown";
        span.setAttribute("http.user_agent", userAgent);
      }
      throw error;
    } finally {
      const duration = Date.now() - startTime;
      span.setAttribute("http.request_duration", duration);
      span.end();
    }
  });
};
