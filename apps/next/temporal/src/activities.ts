import { Context } from "@temporalio/activity";
import { SpanStatusCode, trace } from "@opentelemetry/api";

export async function purchase(id: string): Promise<string> {
  return await trace
    .getTracer("metal")
    .startActiveSpan("purchase", async (span) => {
      span.setAttributes({ id });
      span.addEvent("purchased", {
        attributes: [id],
      });
      span.end();
      return Context.current().info.activityId;
    });
}
