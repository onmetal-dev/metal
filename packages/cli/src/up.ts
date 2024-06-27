import chalk from "chalk";
import { Command } from "commander";
import FormData from "form-data";
import walk from "ignore-walk";
import { createReadStream } from "node:fs";
import { create as createTar } from "tar";
import tmp from "tmp";
import { promptForApp, promptForEnvironment } from "./prompts";
import {
  getAppLink,
  log,
  mustAppByName,
  mustEnvironmentByName,
  mustMetalClient,
  resolveTeamId,
  resolveWorkingDir,
  teamIdOption,
  type Config,
} from "./shared";

export default function up(program: Command, config: Config) {
  program
    .description("Deploy an application")
    .addOption(teamIdOption())
    .option(
      "-e, --environment [environment]",
      "Environment name to deploy to. If not specified, will prompt interactively."
    )
    .option(
      "-a, --app [app]",
      "App name to deploy. Defaults to the linked app."
    )
    .argument(
      "[path]",
      "The absolute or relative path of the directory you want to deploy. Defaults to the current directory."
    )
    .action(async (pathArg, options) => {
      const metal = mustMetalClient(config);
      log(chalk.green(`Collating files to deploy...`));
      const workingDir = resolveWorkingDir(pathArg);
      const appLink = getAppLink(workingDir, config);
      const teamId: string = await resolveTeamId({
        options,
        config,
        appLink,
      });
      const appId: string =
        (options.app
          ? (await mustAppByName(metal, teamId, options.app)).id
          : appLink?.appId) ??
        (
          await promptForApp(metal, {
            userId: config.whoami!.user.id,
            teamId,
            create: true,
          })
        ).id;
      const envId: string = options.environment
        ? (await mustEnvironmentByName(metal, teamId, options.environment)).id
        : (await promptForEnvironment(metal, { teamId, create: true })).id;
      const pathsToCompress = await walk({
        path: workingDir,
        ignoreFiles: [".gitignore", ".dockerignore"],
        includeEmpty: true,
        follow: false,
      });
      if (!pathsToCompress.length) {
        console.error(
          chalk.red(
            "Error: the list of files to compress is empty. Please check that your working directory is nonempty."
          )
        );
        process.exit(1);
      }

      log(chalk.green(`Compressing files...`));
      const tmpTarFile = tmp.fileSync();
      await createTar(
        {
          gzip: true,
          cwd: workingDir,
          file: tmpTarFile.name,
        },
        pathsToCompress
      );

      log(chalk.green(`Uploading...`));
      const form = new FormData();
      form.append("archive", createReadStream(tmpTarFile.name));
      form.append("envId", envId);
      form.append("appId", appId);
      form.append("teamId", teamId);
      const response: Response = await metal.up
        .create({
          archive: createReadStream(tmpTarFile.name),
          teamId,
          envId,
          appId,
        })
        .asResponse();
      if (!response.body) {
        throw new Error("No body found in response");
      }
      const reader = await response.body.getReader();
      const decoder = new TextDecoder();
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          break;
        }
        process.stdout.write(decoder.decode(value));
      }
      tmpTarFile.removeCallback();
      log(chalk.green("Finished!"));
    });
}
