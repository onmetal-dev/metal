import os from "os";

export const serviceName =
  `metal-next` + (process.env.NODE_ENV === "development" ? "-dev" : "");

// queueNameForEnv is the queue that the worker should attach itself to and the
// queue that should be used by application code when submitting workflows.
// This goal is to avoid dev envs and prod env from interacting.
// In the future there should be a dedicated temporal namespace for prod.
export function queueNameForEnv(env: string) {
  const username = os.userInfo().username;
  switch (env) {
    case "development":
      return `metaldev_${username}`;
    case "test":
      return `metaltest_${username}`;
    case "production":
      return "metalprod";
    default:
      throw new Error(`Unknown environment: ${env}`);
  }
}

export const hetznerRedHex = "#D50C2D";

// This matches the foreground color of the app's dark theme that's declared in this file:
// apps/next/app/globals.css
export const whiteishHex = "#F8FAFC";
