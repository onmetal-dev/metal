import { $ } from "bun";
import { inngest } from "./client";

export const helloWorld = inngest.createFunction(
  { id: "hello-world" },
  { event: "test/hello-world" },
  async ({ event, step }) => {
    const { stdout } = await $`echo hello, ${event.data.message}`.quiet();
    await step.sleep("wait-a-moment", "1s");
    return { event, body: stdout.toString().trim() };
  }
);
