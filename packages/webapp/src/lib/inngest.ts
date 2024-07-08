import type { Events } from "@metal/worker/inngest/events";
import { EventSchemas, Inngest } from "inngest";

// Create a client to send and receive events
export const inngest = new Inngest({
  id: "metal",
  schemas: new EventSchemas().fromRecord<Events>(),
});

async function getRuns(eventId: string) {
  const response = await fetch(
    `${process.env.INNGEST_BASE_URL}/v1/events/${eventId}/runs`,
    {
      headers: {
        Authorization: `Bearer ${process.env.INNGEST_SIGNING_KEY}`,
      },
    }
  );
  const json = await response.json();
  return json.data;
}

export async function getRunOutput(eventId: string) {
  let runs = await getRuns(eventId);
  while (runs[0].status !== "Completed") {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    runs = await getRuns(eventId);
    if (runs[0].status === "Failed" || runs[0].status === "Cancelled") {
      throw new Error(`Function run ${runs[0].status}`);
    }
  }
  return runs[0];
}
