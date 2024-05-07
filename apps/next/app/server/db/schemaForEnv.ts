import os from "os";

export default function sqlSchemaForEnv(env: string, ciEnvironment?: string) {
  if (ciEnvironment === "true") {
    return 'metaldev_githubaction';
  }

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
