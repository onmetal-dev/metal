import { registerOTel } from "@vercel/otel";
import { serviceName } from "@/lib/constants";

export function register() {
  registerOTel({
    serviceName,
    autoDetectResources: true,
  });
}
