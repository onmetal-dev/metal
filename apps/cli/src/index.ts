import { Command } from "commander";
import path from "path";
import os from "os";
import chalk from "chalk";
import opener from "opener";
import inquirer from "inquirer";
import z from "zod";
import { readFileSync, existsSync, writeFileSync } from "fs";

// configSchema is the schema of the config file
const configSchema = z.object({
  user: z
    .object({
      id: z.string(),
      email: z.string(),
    })
    .optional(),
  token: z.string().optional(),
});
type Config = z.infer<typeof configSchema>;

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

const METAL_URL = process.env.METAL_URL || "https://www.onmetal.dev";

const program = new Command();
const log = console.log;

// barebones API client that just manages base url and auth
// in the future we probably want to make a formal API client via
// something like https://www.stainlessapi.com/
class MetalApiClient {
  private token: string;
  private baseUrl: string;
  constructor({ token, baseUrl }: { token: string; baseUrl?: string }) {
    this.token = token;
    this.baseUrl = baseUrl ?? METAL_URL;
  }
  async _makeRequest(
    method: "POST" | "GET",
    path: string,
    body?: any
  ): Promise<Response> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers: {
        Authorization: `Bearer ${this.token}`,
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    return response;
  }
  async whoami() {
    const response: Response = await this._makeRequest(
      "GET",
      "/api/user/whoami"
    );
    if (response.status === 401) {
      throw new Error("Token is not valid, please logout/login again");
    }
    return response.json();
  }
}

program
  .name("metal")
  .description("CLI for interacting with onmetal.dev")
  .version("0.0.1", "-v, --version", "output the current version");

program
  .command("whoami")
  .description("Log information about the logged in user")
  .action(async () => {
    if (!config.token) {
      log(`Not logged in. Login with ${chalk.red("metal login")}`);
      return;
    }
    const client = new MetalApiClient({ token: config.token });
    const whoami = await client.whoami();
    log(JSON.stringify(whoami, null, 2));
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
    config.token = undefined;
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
      // basic idea here is:
      // 1. start a server on a random port listening for the redirect from the login page
      // 2. open a browser to the login url
      // 3. wait for the user to login
      // 4. the server will receive the token and save it to the config
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
        opener(url);
      });
    }
    if (!token) {
      log("Unexpected: did not receive token");
      process.exit(1);
    }
    const client = new MetalApiClient({ token: token });

    // make GET request to METAL_URL + /api/user/whoami passing the token as Authorization header
    const whoami = await client.whoami();

    config.user = whoami.user;
    config.token = whoami.token;
    log(`successfully logged in as ${chalk.green(config.user!.email)}`);
    process.exit(0);
  });

program.parse();
