import os from "os";

export default function sqlSchemaForEnv(
  env: "development" | "production" | "test"
) {
  const username = os.userInfo().username;
  switch (env) {
    case "development":
      return `metaldev_${username}`;
    case "test":
      return `metaltest_${username}`;
    case "production":
      return "metalprod";
  }
}
