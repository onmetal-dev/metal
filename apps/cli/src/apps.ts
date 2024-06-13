import type { Application } from "@onmetal/node/resources/applications.mjs";
import type { Team } from "@onmetal/node/resources/teams.mjs";
import chalk from "chalk";
import { Command } from "commander";
import { table } from "table";
import uuidBase62 from "uuid-base62";
import { promptForApp, promptForTeam } from "./prompts";
import {
  log,
  logOutput,
  mustMetalClient,
  outputOption,
  type Config,
} from "./shared";

export default function apps(program: Command, config: Config) {
  program.description("Manage applications");

  program
    .command("list")
    .description("List applications")
    .addOption(outputOption())
    .action(async (options) => {
      const metal = mustMetalClient(config);
      const applications = await metal.applications.list();
      // populate the team object for each application, minimizing the number of requests
      const teamCache = new Map<string, Team>();
      type ApplicationWithTeam = Application & { team: Team };
      const applicationsWithTeam: ApplicationWithTeam[] = await Promise.all(
        applications.map(async (app): Promise<ApplicationWithTeam> => {
          if (!teamCache.has(app.teamId)) {
            const team = await metal.teams.retrieve(app.teamId);
            if (!team) {
              throw new Error(`Team ${app.teamId} not found`);
            }
            teamCache.set(app.teamId, team!);
          }
          return { ...app, team: teamCache.get(app.teamId)! };
        })
      );
      logOutput({
        data: applicationsWithTeam,
        format: options.output,
        humanReadableString: table([
          ["ID", "Name", "Team"],
          ...applicationsWithTeam.map((app) => [
            app.id,
            app.name,
            app.team.name,
          ]),
        ]),
      });
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
      const metal = mustMetalClient(config);
      if (!config.whoami?.user.id) {
        log(`Unknown login state, please logout and login again`);
        return;
      }
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
      const metal = mustMetalClient(config);
      if (!config.whoami?.user.id) {
        log(`Unknown login state, please logout and login again`);
        return;
      }
      const team =
        (options.teamId && (await metal.teams.retrieve(options.teamId))) ||
        (await promptForTeam(metal));
      const app =
        (options.appId && (await metal.applications.retrieve(options.appId))) ||
        (await promptForApp(metal, {
          teamId: team.id,
          userId: config.whoami.user.id,
        }));
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
