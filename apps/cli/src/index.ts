import { Command } from "commander";
import path from "path";
import os from "os";
import chalk from "chalk";
import opener from "opener";
import inquirer from "inquirer";
import z from "zod";
import { readFileSync, existsSync, writeFileSync } from "fs";

// define a zod schema for a config file. Has a user top level field with a token for the user
const userSchema = z.object({
  id: z.string(),
  token: z.string(),
  email: z.string(),
});
type User = z.infer<typeof userSchema>;
const configSchema = z.object({
  user: userSchema.optional(),
});
type Config = z.infer<typeof configSchema>;

// setup / load config
const configPath = path.join(os.homedir(), ".config", "metal", "config.json");
if (!existsSync(configPath)) {
  writeFileSync(configPath, "{}", "utf8");
}
const config: Config = JSON.parse(readFileSync(configPath, "utf8"));
// on exit, write the config
process.on("exit", () => {
  writeFileSync(configPath, JSON.stringify(config, null, 2), "utf8");
});

const METAL_URL = process.env.METAL_URL || "https://www.onmetal.dev";

const program = new Command();
const log = console.log;

program
  .name("metal")
  .description("CLI for interacting with onmetal.dev")
  .version("0.0.1", "-v, --version", "output the current version");

program
  .command("logout")
  .description("Log out from onmetal.dev")
  .action(async () => {
    const { user } = config;
    if (!user) {
      log(`Not logged in, so you're already logged out :)`);
      process.exit(0);
    }
    config.user = undefined;
    log("Logged out");
    process.exit(0);
  });

program
  .command("login")
  .description("Login to onmetal.dev")
  .option("--token", "or METAL_TOKEN. Provide a token manually, useful for CI")
  .action(async (str, options) => {
    if (config.user) {
      log(
        `Already logged in as ${config.user.email}. Use ${chalk.red(
          "metal logout"
        )} to log out.`
      );
      process.exit(0);
    }
    let token = process.env.METAL_TOKEN;
    if (options.token) {
      token = options.token;
    }
    if (!token) {
      const port = Math.floor(Math.random() * 10000) + 50000;
      const url = `${METAL_URL}/login-to-cli?next=http%3A%2F%2Flocalhost%3A${port}%2F`;
      const answers = await inquirer.prompt({
        name: "continue",
        type: "confirm",
        message: "Ready to open the browser?",
      });
      if (!answers.continue) {
        log("Exiting");
        process.exit(0);
      }
      token = await new Promise((resolve) => {
        Bun.serve({
          port,
          fetch: async (req) => {
            const url = new URL(req.url);
            token = url.searchParams.get("token") ?? undefined;
            if (!token) {
              throw new Error("No token found in url");
            }
            setTimeout(() => {
              resolve(token); // need to give time for the redirect response to be delivered
            }, 500);
            return Response.redirect(`${METAL_URL}/login-to-cli-success`, 302);
          },
        });
        opener(url);
      });
    }
    if (!token) {
      log("Unexpected: did not receive token");
      process.exit(1);
    }

    // make GET request to METAL_URL + /api/user/whoami passing the token as Authorization header
    const whoamiResponse: Response = await fetch(
      `${METAL_URL}/api/user/whoami`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    const body = await whoamiResponse.json();

    // insert user id, token, email
    const { user } = body;
    let email = "";
    if (
      user.emailAddresses &&
      user.emailAddresses.length > 0 &&
      user.emailAddresses[0].emailAddress
    ) {
      email = user.emailAddresses[0].emailAddress;
    } else {
      console.error("No email address found for user.");
      process.exit(1);
    }
    config.user = {
      id: user.id,
      token,
      email,
    };
    log(`successfully logged in as ${chalk.green(email)}`);
    process.exit(0);
  });

program.parse();
