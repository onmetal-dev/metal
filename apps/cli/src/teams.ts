import { NotFoundError } from "@onmetal/node";
import { Command } from "commander";
import { table } from "table";
import {
  log,
  logOutput,
  mustMetalClient,
  outputOption,
  type Config,
} from "./shared";

export default function teams(program: Command, config: Config) {
  program.description("Manage teams");

  program
    .command("get")
    .description("Get team details")
    .addOption(outputOption())
    .argument("<teamId>", "Team ID")
    .action(async (teamId, options) => {
      const metal = mustMetalClient(config);
      try {
        const team = await metal.teams.retrieve(teamId);
        logOutput({
          data: team,
          format: options.output,
          humanReadableString: table([
            ["ID", "Name", "Created"],
            [team.id, team.name, team.createdAt],
          ]),
        });
      } catch (error: any) {
        if (error instanceof NotFoundError) {
          log(`Team ${teamId} not found`);
          return;
        }
        throw error;
      }
    });

  program
    .command("list")
    .description("List all teams")
    .addOption(outputOption())
    .action(async (options) => {
      const metal = mustMetalClient(config);
      const teams = await metal.teams.list({});
      logOutput({
        data: teams,
        format: options.output,
        humanReadableString: table([
          ["ID", "Name", "Created"],
          ...teams.map((team) => [team.id, team.name, team.createdAt]),
        ]),
      });
    });

  // todo: need a team create endpoint in the API
  // program
  //   .command("create")
  //   .description("Create a new team")
  //   .requiredOption("-n, --name <name>", "Team name")
  //   .option("-d, --description <description>", "Team description")
  //   .action(async (options) => {
  //     const metal = mustMetalClient(config);
  //     const id = uuidBase62.v4();
  //     const { name, description } = options;
  //     const team = await metal.teams.create(id, {
  //       id,
  //       name,
  //       description,
  //     });
  //     log(chalk.green(`Team ${team.id} created successfully`));
  //   });
}
