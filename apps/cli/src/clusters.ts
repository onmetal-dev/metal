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
  program.description("Manage Hetzner clusters created by your team");

  program
    .command("get")
    .description("Get cluster details")
    .argument("<clusterId>", "Cluster ID")
    .action(async (clusterId) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      try {
        const cluster = await metal.hetznerClusters.retrieve(clusterId);
        log(cluster);
      } catch (error: any) {
        if (error instanceof NotFoundError) {
          log(`Cluster ${clusterId} not found`);
          return;
        }
        throw error;
      }
    });
  program
    .command("list")
    .description("List Hetzner clusters for your teams")
    .action(async () => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const clusters = await metal.hetznerClusters.list();
      log(JSON.stringify(clusters, null, 2));
    });

  program
    .command("create")
    .description("Create a Hetzner cluster for your team")
    .option(
      "-t, --team [teamId]",
      "Team ID (optional). Defaults to first team."
    )
    .requiredOption(
      "-l, --location [location]",
      "Location: fsn1, nbg1, or hel1"
    )
    .requiredOption(
      "-n, --instance-type [nodeType]",
      "Instance type: cax11, cax21, cax31, or cax41"
    )
    .option("-c, --node-count [nodeCount]", "Number of nodes to create", "1")
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
      const id = uuidBase62.v4();
      const teamId = options.teamId || config.whoami.teams[0].id;
      const { location, instanceType } = options;
      const nodeCount = parseInt(options.nodeCount);
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      const cluster = await metal.hetznerClusters.create(id, {
        id,
        teamId,
        location,
        nodeGroups: [
          {
            instanceType,
            maxNodes: nodeCount,
            minNodes: nodeCount,
            type: "all",
          },
        ],
      });
      log(chalk.green(`Cluster ${cluster.id} created successfully`));
    });
  program
    .command("delete")
    .description("Delete cluster")
    .argument("<clusterId>", "Cluster ID")
    .action(async (clusterId) => {
      if (!config.whoami) {
        log(`Not logged in. Login with ${chalk.red("metal login")}`);
        return;
      }
      const metal = new Metal({ baseURL, metalAPIKey: config.whoami.token });
      await metal.hetznerClusters.delete(clusterId);
      log(chalk.green(`Cluster ${clusterId} deleted successfully`));
    });
}
