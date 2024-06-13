import { Command } from "commander";

import { NotFoundError } from "@onmetal/node";
import type { HetznerCluster } from "@onmetal/node/resources/hetzner-clusters.mjs";
import type { Team } from "@onmetal/node/resources/teams.mjs";
import chalk from "chalk";
import { table } from "table";
import uuidBase62 from "uuid-base62";
import {
  getAppLink,
  log,
  logOutput,
  mustMetalClient,
  outputOption,
  resolveTeamId,
  teamIdOption,
  type Config,
} from "./shared";

export default function projects(program: Command, config: Config) {
  program.description("Manage Hetzner clusters created by your team");

  program
    .command("get")
    .description("Get cluster details")
    .argument("<clusterId>", "Cluster ID")
    .action(async (clusterId) => {
      const metal = mustMetalClient(config);
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
    .description("List Hetzner clusters for your team")
    .addOption(outputOption())
    .addOption(teamIdOption())
    .action(async (options) => {
      const teamId: string | undefined = await resolveTeamId({
        options,
        config,
        appLink: getAppLink(process.cwd(), config),
      });
      const metal = mustMetalClient(config);
      const clusters = (await metal.hetznerClusters.list({})).filter(
        (cluster) => teamId === undefined || cluster.teamId === teamId
      );
      const teamCache = new Map<string, Team>();
      type ClusterWithTeam = HetznerCluster & { team: Team };
      const clustersWithTeam: ClusterWithTeam[] = await Promise.all(
        clusters.map(async (cluster): Promise<ClusterWithTeam> => {
          if (!teamCache.has(cluster.teamId)) {
            const team = await metal.teams.retrieve(cluster.teamId);
            if (!team) {
              throw new Error(`Team ${cluster.teamId} not found`);
            }
            teamCache.set(cluster.teamId, team!);
          }
          return { ...cluster, team: teamCache.get(cluster.teamId)! };
        })
      );

      logOutput({
        data: clusters,
        format: options.output,
        humanReadableString: table([
          [
            "ID",
            "Name",
            "Location",
            "Min Nodes",
            "Max Nodes",
            "Team ID",
            "Team",
            "Status",
          ],
          ...clustersWithTeam.map((cluster) => [
            cluster.id,
            cluster.name,
            cluster.location,
            cluster.nodeGroups.reduce(
              (acc, nodeGroup) => acc + nodeGroup.minNodes,
              0
            ),
            cluster.nodeGroups.reduce(
              (acc, nodeGroup) => acc + nodeGroup.maxNodes,
              0
            ),
            cluster.teamId,
            cluster.team.name,
            cluster.status,
          ]),
        ]),
      });
    });

  program
    .command("create")
    .description("Create a Hetzner cluster for your team")
    .addOption(teamIdOption())
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
      const metal = mustMetalClient(config);
      if (!config.whoami?.teams) {
        log(`Unknown login state, please logout and login again`);
        return;
      }
      if (config.whoami.teams.length != 1) {
        log(
          `You are a member of ${config.whoami.teams.length} teams. Please specify a team with --team`
        );
        return;
      }
      const id = uuidBase62.v4();
      const teamId = await resolveTeamId({
        options,
        config,
        appLink: getAppLink(process.cwd(), config),
      });
      const { location, instanceType } = options;
      const nodeCount = parseInt(options.nodeCount);
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
      const metal = mustMetalClient(config);
      await metal.hetznerClusters.delete(clusterId);
      log(chalk.green(`Cluster ${clusterId} deleted successfully`));
    });
}
