import { Command } from "commander";
import path from "path";
import os from "os";
import chalk from "chalk";
import opener from "opener";
import inquirer from "inquirer";
import Metal from "@onmetal/node";
import { type Config } from "./types";
import { readFileSync, existsSync, writeFileSync, mkdirSync } from "fs";
import { promisify } from "node:util";
import { exec as execCallbackBased } from "node:child_process";
import { create as createTar } from "tar";
import { request as insecureRequest } from "node:http";
import { request as secureRequest } from "node:https";

const exec = promisify(execCallbackBased);

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
// write the config on exit
process.on("exit", () => {
  writeFileSync(configPath, JSON.stringify(config, null, 2), "utf8");
});

const baseURL = process.env.METAL_BASE_URL || "https://www.onmetal.dev/api";
const baseUrlObj = new URL(baseURL);
const baseDomainWithProtocol = `${baseUrlObj.protocol}//${baseUrlObj.host}`;

const isLocalhost =
  baseUrlObj.hostname === "localhost" || baseUrlObj.hostname === "127.0.0.1";
const nodeRequest = isLocalhost ? insecureRequest : secureRequest;

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
        // TODO MET-10: This was last updated 4 years ago. We should find an alternative?
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

const checkUserConfig = () => {
  if (!config.whoami) {
    log(`Oops! You're not logged in :)`);
    process.exit(1);
  }

  return config as Required<Config>;
};

program
  .command("up")
  .description("Deploy a project")
  .option(
    "--token",
    "Manually provide a Metal token, or set the METAL_TOKEN environment variable. Useful for CI."
  )
  .action(async (str, options) => {
    let step = 1;
    log(`[${step}] Checking for token...`);
    const userConfig = checkUserConfig();
    // Token hierachy: commandline > config file > environment variable
    const token =
      options.token || userConfig.whoami.token || process.env.METAL_TOKEN;
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

    log(`[${++step}] Compressing files...`);
    const payloadStream = createTar(
      {
        gzip: true,
        cwd: process.cwd(),
      },
      pathsToArchive
    );

    log(`[${++step}] Uploading...`);
    const reqOptions = {
      host: baseUrlObj.hostname,
      port: baseUrlObj.port || 80,
      path: "/api/deploy/up",
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        Authorization: `Bearer ${token}`,
        Accept: "application/json",
      },
    };

    const bodyAsString = await new Promise<string>((resolve, reject) => {
      const request = nodeRequest(reqOptions, (response) => {
        let bodyJSONString = "";
        response.on("data", (chunk) => {
          bodyJSONString += chunk;
        });

        response.on("error", (err) => {
          console.error(`Problem with response from upload: ${err.message}`);
          reject(err);
        });

        response.on("end", () => {
          resolve(bodyJSONString);
        });
      });

      request.on("error", (err) => {
        console.error(`Problem with upload request: ${err.message}`);
        reject(err);
      });

      // Write compressed data to request's body
      payloadStream.pipe(request);
    });

    const body = JSON.parse(bodyAsString);
    log(`--> Deployment started. Tag is ${body.tag}`);

    log(`[${++step}] Checking deployment status...`);
    const statusPromise = new Promise<string>((resolve, reject) => {
      const statusRequest = nodeRequest(
        `${baseDomainWithProtocol}/api/deploy/${body.tag}/status`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/json",
          },
        }
      );

      statusRequest.on("response", (res) => {
        res.on("data", (chunk) => {
          log(`[${++step}] ${chunk.toString()}`);
        });
        res.on("error", (err) => {
          console.error("Failed to read response.");
          console.error(err);
          reject(err);
        });
        res.on("end", () => {
          resolve("Deployment finished.");
        });
      });

      statusRequest.on("error", (err) => {
        console.error(`Error in status request: ${err.message}`);
        reject(err);
      });

      statusRequest.end();
    });

    const result = await statusPromise;
    log(`[END] ${result}`);
  });

import projects from "./projects";
projects(program.command("projects"), config, baseURL);
import clusters from "./clusters";
clusters(program.command("clusters"), config, baseURL);

program.parse();
