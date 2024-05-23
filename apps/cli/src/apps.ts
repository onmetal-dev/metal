import { Command } from "commander";

import Metal from "@onmetal/node";
import type { Application } from "@onmetal/node/resources/applications.mjs";
import type { Team } from "@onmetal/node/resources/teams.mjs";
import chalk from "chalk";
import inquirer from "inquirer";
import uuidBase62 from "uuid-base62";
import { type Config } from "./types";

const log = console.log;

export default function apps(
  program: Command,
  config: Config,
  baseURL: string
) {
  program.description("Manage applications");

  program
    .command("list")
    .description("List applications")
    .action(async () => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const applications = await metal.applications.list();
      log(JSON.stringify(applications, null, 2));
    });

  program
    .command("create")
    .description("Create an application")
    .argument("<name>", "Application name. Required.")
    .option(
      "-t, --team [teamId]",
      "Team ID (optional). Will prompt if not specified."
    )
    .action(async (name, options) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({
        baseURL,
        metalAPIKey: config.whoami.token,
      });

      const team =
        (options.teamId && (await metal.teams.retrieve(options.teamId))) ||
        (await promptForTeam(metal));
      const id = uuidBase62.v4();
      const application = await metal.applications.create(id, {
        id,
        teamId: team.id,
        creatorId: config.whoami.user.id,
        name: name,
      });
      log(chalk.green(`Application ${application.id} created successfully`));
    });

  program
    .command("link")
    .description("Link a directory with an application")
    .argument(
      "[directory]",
      "Directory to link (optional). Defaults to current directory."
    )
    .option(
      "-t, --team [teamId]",
      "Team ID (optional). Will prompt if not specified."
    )
    .option(
      "-a, --app [appId]",
      "Application to link to (optional). Will prompt if not specified."
    )
    .action(async (directory, options) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const team =
        (options.teamId && (await metal.teams.retrieve(options.teamId))) ||
        (await promptForTeam(metal));
      const app =
        (options.appId && (await metal.applications.retrieve(options.appId))) ||
        (await promptForApp(metal, team));
      directory = directory || process.cwd();
      config.appLinks = config.appLinks || {};
      config.appLinks[directory] = {
        directory,
        appId: app.id,
        appName: app.name,
        teamId: team.id,
        teamName: team.name,
      };
      console.log(
        chalk.green(`Linked directory ${directory} to application ${app.name}.`)
      );
    });
  program
    .command("unlink")
    .description("Unlink a directory from an application")
    .argument(
      "[directory]",
      "Directory to unlink (optional). Defaults to current directory."
    )
    .action(async (directory) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      directory = directory || process.cwd();
      if (config.appLinks === undefined || !config.appLinks[directory]) {
        log(chalk.red(`No link found for directory ${directory}`));
        return;
      }
      log(
        chalk.green(
          `Unlinked directory ${directory} from application ${config.appLinks[directory].appName}`
        )
      );
      delete config.appLinks[directory];
    });
}

async function promptForTeam(metal: Metal): Promise<Team> {
  const teams = await metal.teams.list();
  if (teams.length === 0) {
    log(chalk.red("You are not a member of any teams!"));
    process.exit(1);
  }
  const answers = await inquirer.prompt({
    type: "list",
    name: "team",
    message: "Please select a team",
    choices: teams.map((team) => team.name),
  });
  if (!answers || !answers.team) {
    log("Exiting");
    process.exit(1);
  }
  const team = teams.find((team) => team.name === answers.team);
  if (!team) {
    log("Exiting");
    process.exit(1);
  }
  return team;
}

async function promptForApp(metal: Metal, team: Team): Promise<Application> {
  const apps = (await metal.applications.list()).filter(
    (app) => app.teamId === team.id
  );
  if (apps.length === 0) {
    log(chalk.red(`No apps found for team '${team.name}'`));
    log(chalk.red("Create an app with `metal apps create`"));
    process.exit(1);
  }
  const answers = await inquirer.prompt({
    type: "list",
    name: "app",
    message: "Please select an app",
    choices: apps.map((app) => app.name),
  });
  if (!answers || !answers.app) {
    log("Exiting");
    process.exit(1);
  }
  const app = apps.find((app) => app.name === answers.app);
  if (!app) {
    log("Exiting");
    process.exit(1);
  }
  return app;
}
