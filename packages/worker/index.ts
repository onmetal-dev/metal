import { serve } from "inngest/bun";
import { client, functions } from "./inngest";

const port = process.env.PORT ?? 3010;
console.log("Serving on port", port);
Bun.serve({
  port,
  fetch(request: Request) {
    const url = new URL(request.url);

    if (url.pathname === "/api/inngest") {
      return serve({ client, functions })(request);
    }

    return new Response("Not found", { status: 404 });
  },
});
