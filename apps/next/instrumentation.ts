import { registerOTel } from "@vercel/otel";

// for some god awful reason you have to do imports inside
// the register function
// https://github.com/vercel/next.js/issues/49565

export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { serviceName } = await import("@/lib/constants");
    registerOTel({
      serviceName,
      autoDetectResources: true,
    });
  }
}
