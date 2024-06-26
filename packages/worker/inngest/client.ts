import { EventSchemas, Inngest } from "inngest";
import type { Events } from "./events";

// Create a client to send and receive events
export const inngest = new Inngest({
  id: "metal",
  schemas: new EventSchemas().fromRecord<Events>(),
});
