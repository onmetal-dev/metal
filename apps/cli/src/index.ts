import { Command } from "commander";
import path from "path";
import os from "os";
import chalk from "chalk";
import opener from "opener";
import inquirer from "inquirer";
import { readFileSync, existsSync, writeFileSync } from "fs";
import Metal from "@onmetal/node";
import { type WhoAmI } from "@onmetal/node/resources/whoami.mjs";

interface Config {
  whoami?: WhoAmI;
}
// setup / load config
const configPath = path.join(os.homedir(), ".config", "metal", "config.json");
if (!existsSync(configPath)) {
  writeFileSync(configPath, "{}", "utf8");
}
const config: Config = JSON.parse(readFileSync(configPath, "utf8"));
// write the config on exit
process.on("exit", () => {
  writeFileSync(configPath, JSON.stringify(config, null, 2), "utf8");
});

const baseURL = process.env.METAL_BASE_URL || "https://www.onmetal.dev/api";
const baseUrlObj = new URL(baseURL);
const baseDomainWithProtocol = `${baseUrlObj.protocol}//${baseUrlObj.host}`;

const program = new Command();
const log = console.log;

program
  .name("metal")
  .description("CLI for interacting with onmetal.dev")
  .version("0.0.1", "-v, --version", "output the current version");

program
  .command("whoami")
  .description("Log information about the logged in user")
  .action(async () => {
    if (!config.whoami) {
      log(`Not logged in. Login with ${chalk.red("metal login")}`);
      return;
    }
    const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
    const whoAmI = await metal.whoami.retrieve();
    log(JSON.stringify(whoAmI, null, 2));
  });

program
  .command("logout")
  .description("Log out from onmetal.dev")
  .action(async () => {
    const user = config.whoami?.user;
    if (!user) {
      log(`Not logged in, so you're already logged out :)`);
      process.exit(0);
    }
    config.whoami = undefined;
    log("Logged out");
  });

program
  .command("login")
  .description("Login to onmetal.dev")
  .option("--token", "or METAL_TOKEN. Provide a token manually, useful for CI")
  .action(async (str, options) => {
    if (config.whoami) {
      log(
        `Already logged in as ${config.whoami.user.email}. Use ${chalk.red(
          "metal logout"
        )} to log out.`
      );
      return;
    }
    let token = process.env.METAL_TOKEN;
    if (options.token) {
      token = options.token;
    }
    if (!token) {
      // basic idea here is:
      // 1. start a server on a random port listening for the redirect from the login page
      // 2. open a browser to the login url
      // 3. wait for the user to login
      // 4. the server will receive the token and save it to the config
      const port = Math.floor(Math.random() * 10000) + 50000;
      const url = `${baseDomainWithProtocol}/login-to-cli?next=http%3A%2F%2Flocalhost%3A${port}%2F`;
      const answers = await inquirer.prompt({
        name: "continue",
        type: "confirm",
        message: "Ready to open the browser?",
      });
      if (!answers.continue) {
        log("Exiting");
        return;
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
            return Response.redirect(
              `${baseDomainWithProtocol}/login-to-cli-success`,
              302
            );
          },
        });
        opener(url);
      });
    }
    if (!token) {
      log("Unexpected: did not receive token");
      process.exit(1);
    }
    const metal = new Metal({ baseURL, metalAPIKey: token });
    const whoami = await metal.whoami.retrieve();

    config.whoami = whoami;
    log(`successfully logged in as ${chalk.green(config.whoami.user!.email)}`);
    process.exit(0);
  });

program.parse();
