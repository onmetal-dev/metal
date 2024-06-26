import { serve } from "inngest/bun";
import { client, functions } from "./inngest";

Bun.serve({
  port: 3010,
  fetch(request: Request) {
    const url = new URL(request.url);

    if (url.pathname === "/api/inngest") {
      return serve({ client, functions })(request);
    }

    return new Response("Not found", { status: 404 });
  },
});
