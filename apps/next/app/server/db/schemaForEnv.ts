import os from "os";

// NOTE: the CI environment variable is set by GitHub Actions. More info:
// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
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
