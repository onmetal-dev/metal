import os from "os";

export default function sqlSchemaForEnv(env: string) {
  const username = os.userInfo().username;
  switch (env) {
    case "ci":
    case "development":
    case "test":
      return `metaltest_${username}`;

    case "production":
      return "metalprod";

    default:
      throw new Error(`Unknown environment: ${env}`);
  }
}
