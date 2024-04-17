import { Command } from "commander";
import path from "path";
import os from "os";
import chalk from "chalk";
import opener from "opener";
import inquirer from "inquirer";
import z from "zod";
import { readFileSync, existsSync, writeFileSync, mkdirSync } from "fs";
import { promisify } from "node:util";
import { exec as execCallbackBased } from "node:child_process";
import { create as createTar } from "tar";

const exec = promisify(execCallbackBased);

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
const configParentDir = path.join(os.homedir(), ".config", "metal");
const configPath = path.join(configParentDir, "config.json");
if (!existsSync(configParentDir)) {
  mkdirSync(configParentDir);
}
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
  .command("whoami")
  .description("Log information about the logged in user")
  .action(async () => {
    const { user } = config;
    if (!user) {
      log(`Not logged in. Login with ${chalk.red("metal login")}`);
      return;
    }
    // make GET request to METAL_URL + /api/user/whoami passing the token as Authorization header
    const whoamiResponse: Response = await fetch(
      `${METAL_URL}/api/user/whoami`,
      {
        headers: {
          Authorization: `Bearer ${user.token}`,
        },
      }
    );
    if (whoamiResponse.status === 401) {
      log("Token is not valid, please logout/login again");
      return;
    }
    const body = await whoamiResponse.json();
    log(JSON.stringify(body, null, 2));
  });

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
      return;
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
            return Response.redirect(`${METAL_URL}/login-to-cli-success`, 302);
          },
        });
        // TODO MET-10: This was last updated 4 years ago. We should find an alternative?
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

const checkUserConfig = () => {
  if (!config.user) {
    log(`Oops! You're not logged in :)`);
    process.exit(1);
  }

  return config as Required<Config>;
}

program
  .command("up")
  .description("Deploy a project")
  .option("--token", "Manually provide a Metal token, or set the METAL_TOKEN environment variable. Useful for CI.")
  .action(async (str, options) => {
    /* Planned steps
     * - [DONE] use git ls-files to get a list of git-tracked files.
     * -- [DONE] filter out any .gitignore
     * - targz them up
     * - make a POST request to METAL_URL + /api//v1/deploy with:
     * -- the user's token as a bearer token in the Authorization header
     * -- add the config for the current project too somehow so that we know the project we're deploying.
     * -- upload the tarball as a multipart request. Once the upload is complete, return a tag to the command.
     * - then this command calls GET METAL_URL + /api/v1/deploy/{tag} to check the status of the deployment.
     * - if the deployment is ongoing, stream responses to the command line.
     * - if it has already finished with a success or failure, return that and end this command.
     */

    let step = 1;
    log(`[${step}] Checking for token...`);
    const userConfig = checkUserConfig();
    // Token hierachy: commandline > config file > environment variable
    const token = options.token || userConfig.user.token || process.env.METAL_TOKEN;
    if (!token) {
      log("Error! You must configure a Metal API token.");
      process.exit(1);
    }

    log(`[${++step}] Collating files to deploy...`);
    const { stdout, stderr } = await exec(`git ls-files`);
    if (stderr) {
      console.error(stderr);
      process.exit(1);
    }

    const pathsToArchive = stdout
      .split("\n")
      .filter((path) => !!path && !path.endsWith(".gitignore"));

    log(`[${++step}] Archiving files...`);
    const payloadStream = createTar({
      gzip: true,
      cwd: process.cwd(),
    }, pathsToArchive);

    log(`[${++step}] Uploading...`);
    log(`[${++step}] Deployment started. ID is XYZ...`);
  })

program.parse();
