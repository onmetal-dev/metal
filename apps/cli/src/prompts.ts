import Metal from "@onmetal/node";
import { type Application } from "@onmetal/node/resources/applications.mjs";
import { type Environment } from "@onmetal/node/resources/environments.mjs";
import uuidBase62 from "uuid-base62";

import input from "@inquirer/input";
import select from "@inquirer/select";
import { type Team } from "@onmetal/node/resources/teams.mjs";
import chalk from "chalk";
import { log } from "./shared";

export async function promptForTeam(metal: Metal): Promise<Team> {
  const teams = await metal.teams.list();
  if (teams.length === 0) {
    log(chalk.red("You are not a member of any teams!"));
    process.exit(1);
  }
  const answer = await select({
    message: "Please select a team",
    choices: teams.map((team) => ({ value: team.id, name: team.name })),
  });
  if (!answer) {
    log("Exiting");
    process.exit(1);
  }
  const team = teams.find((team) => team.id === answer);
  if (!team) {
    log("Exiting");
    process.exit(1);
  }
  return team;
}

interface PromptForAppOptions {
  userId: string;
  teamId: string;
  create?: boolean;
}

export async function promptForApp(
  metal: Metal,
  options: PromptForAppOptions
): Promise<Application> {
  const { userId, teamId, create } = options;
  const apps = (await metal.applications.list()).filter(
    (app) => app.teamId === teamId // todo: support this in the api
  );
  if (apps.length === 0 && create === false) {
    log(chalk.red(`No apps found for team '${teamId}'`));
    log(chalk.red("Create an app with `metal apps create`"));
    process.exit(1);
  }
  const createPrompt = "Create a new app";
  const appId = await select({
    message: "Please select an app",
    choices: apps
      .map((app) => ({ value: app.id, name: app.name }))
      .concat(create ? { value: createPrompt, name: createPrompt } : []),
  });
  if (!appId) {
    log("Exiting");
    process.exit(1);
  }
  if (appId === createPrompt) {
    const newAppName = await input({
      message: "Enter the name of the app you'd like to create",
      validate: (input) => {
        if (input.length === 0 || !/^[a-zA-Z0-9_-]+$/.test(input)) {
          return "App name must be nonempty and match the regex ^[a-zA-Z0-9_-]+$";
        }
        return true;
      },
    });
    return await metal.applications.create(uuidBase62.v4(), {
      name: newAppName,
      teamId,
      creatorId: userId,
    });
  }

  const app = apps.find((app) => app.id === appId);
  if (!app) {
    log("Exiting");
    process.exit(1);
  }
  return app;
}

interface PromptForEnvironmentOptions {
  teamId: string;
  create?: boolean;
}

export async function promptForEnvironment(
  metal: Metal,
  options: PromptForEnvironmentOptions
): Promise<Environment> {
  const { teamId, create } = options;
  const environments = (await metal.environments.list({ teamId })).filter(
    (e) => e.teamId === teamId
  );
  if (environments.length === 0 && create === false) {
    log(chalk.red(`No environments found for team '${teamId}'`));
    log(chalk.red("Create an environment with `metal environments create`"));
    process.exit(1);
  }
  const createPrompt = "Create a new environment";
  const environmentId = await select({
    message: "Please select an environment",
    choices: environments
      .map((e) => ({ value: e.id, name: e.name }))
      .concat(create ? { value: createPrompt, name: createPrompt } : []),
  });
  if (!environmentId) {
    log("Exiting");
    process.exit(1);
  }
  if (environmentId === createPrompt) {
    const newEnvName = await input({
      message: "Enter the name of the environment you'd like to create",
      validate: (input) => {
        return input.length > 0 && /^[a-zA-Z_-]+$/.test(input);
      },
    });
    return await metal.environments.create(uuidBase62.v4(), {
      name: newEnvName,
      teamId,
    });
  }
  const environment = environments.find((e) => e.id === environmentId);
  if (!environment) {
    log("Exiting");
    process.exit(1);
  }
  return environment;
}
