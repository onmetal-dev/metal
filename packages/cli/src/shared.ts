import Metal from "@onmetal/node";
import type { Application } from "@onmetal/node/resources/applications.mjs";
import type { Environment } from "@onmetal/node/resources/environments.mjs";
import { type WhoAmI } from "@onmetal/node/resources/whoami.mjs";
import chalk from "chalk";
import { Option } from "commander";
import path from "path";
import * as yaml from "yaml";
import { promptForTeam } from "./prompts";

export const log = console.log;

type AppLink = {
  directory: string;
  appId: string;
  appName: string;
  teamId: string;
  teamName: string;
};

export interface Config {
  whoami?: WhoAmI;
  appLinks?: {
    [directory: string]: AppLink;
  };
}

export function getAppLink(directory: string, config: Config): AppLink | null {
  return config.appLinks?.[directory] || null;
}

export async function mustAppByName(
  metal: Metal,
  teamId: string,
  appName: string
): Promise<Application> {
  const apps = await metal.applications.list({});
  const app = apps.find((app) => app.name === appName && app.teamId === teamId);
  if (!app) {
    throw new Error(`No app found with name ${appName} and teamId ${teamId}`);
  }
  return app;
}

export async function mustEnvironmentByName(
  metal: Metal,
  teamId: string,
  environmentName: string
): Promise<Environment> {
  const environments: Environment[] = await metal.environments.list({ teamId });
  const environment = environments.find(
    (e) => e.name === environmentName && e.teamId === teamId
  );
  if (!environment) {
    throw new Error(
      `No environment found with name ${environmentName} and teamId ${teamId}`
    );
  }
  return environment;
}

// mustMetalClient constructs a metal client using process.env.METAL_TOKEN, or if not present, the logged in user's stored token
export function mustMetalClient(config: Config): Metal {
  if (!config.whoami) {
    log(`Not logged in. Login with ${chalk.red("metal login")}`);
    process.exit(1);
  }
  const baseURL = process.env.METAL_BASE_URL || "https://www.onmetal.dev/api";
  return new Metal({
    baseURL,
    metalAPIKey: process.env.METAL_TOKEN || config.whoami.token,
  });
}

export function outputOption() {
  return new Option(
    "-o, --output <format>",
    "Output format: json, yaml, human"
  ).default("human");
}
export interface LogOutputOptions {
  data: any;
  format: string;
  humanReadableString?: string;
}
export function logOutput({
  data,
  format,
  humanReadableString,
}: LogOutputOptions) {
  switch (format) {
    case "yaml":
      log(yaml.stringify(data));
      break;
    case "json":
      log(JSON.stringify(data, null, 2));
      break;
    case "human":
    default:
      log(humanReadableString || data);
      break;
  }
}

export function teamIdOption() {
  return new Option(
    "-t, --teamId [teamId]",
    "Team ID. Defaults to the linked team. If not specified, you will be prompted to select a team."
  );
}

interface ResolveTeamIdOptions {
  options: {
    teamId?: string;
  };
  config: Config;
  appLink: AppLink | null;
}

export async function resolveTeamId({
  options,
  config,
  appLink,
}: ResolveTeamIdOptions): Promise<string> {
  const metal = mustMetalClient(config);
  return options.teamId ?? appLink?.teamId ?? (await promptForTeam(metal)).id;
}

export function resolveWorkingDir(pathArg: string | undefined): string {
  return pathArg
    ? path.isAbsolute(pathArg)
      ? pathArg
      : path.resolve(process.cwd(), pathArg)
    : process.cwd();
}
