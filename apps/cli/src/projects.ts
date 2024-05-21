import { Command } from "commander";

import { type Config } from "./types";
import chalk from "chalk";
import Metal, { NotFoundError } from "@onmetal/node";
import uuidBase62 from "uuid-base62";

const log = console.log;

export default function projects(
  program: Command,
  config: Config,
  baseURL: string
) {
  program.description("Manage Hetzner projects connected to your team");

  program
    .command("get")
    .description("Get project details")
    .argument("<projectId>", "Project ID")
    .action(async (projectId) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      try {
        const project = await metal.hetznerProjects.retrieve(projectId);
        log(project);
      } catch (error: any) {
        if (error instanceof NotFoundError) {
          log(`Project ${projectId} not found`);
          return;
        }
        throw error;
      }
    });
  program
    .command("list")
    .description("List Hetzner projects connected to your teams")
    .action(async () => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const projects = await metal.hetznerProjects.list();
      log(projects);
    });
  program
    .command("create")
    .description("Connect a Hetzner project to your team")
    .option("-i, --id [id]", "Project ID")
    .option("-t, --team [teamId]", "Team ID")
    .requiredOption("-n, --hetznerName <hetznerName>", "Hetzner project name")
    .requiredOption("-a, --apiToken <apiToken>", "Hetzner API token")
    .action(async (options) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      if (config.whoami.teams.length != 1) {
        log(
          `You are a member of ${config.whoami.teams.length} teams. Please specify a team with --team`
        );
        return;
      }
      const id = options.id || uuidBase62.v4();
      const teamId = options.teamId || config.whoami.teams[0].id;
      const { hetznerName, apiToken } = options;
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const project = await metal.hetznerProjects.create(id, {
        id,
        hetznerName,
        teamId,
        creatorId: config.whoami.user.id,
        hetznerApiToken: apiToken,
      });
      log(chalk.green(`Project ${project.id} created successfully`));
    });
  program
    .command("delete")
    .description("Delete project")
    .argument("<projectId>", "Project ID")
    .action(async (projectId) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      await metal.hetznerProjects.delete(projectId);
      log(chalk.green(`Project ${projectId} deleted successfully`));
    });
}
