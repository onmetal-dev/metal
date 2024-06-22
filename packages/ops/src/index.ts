import { Command } from "commander";
import mgmtCluster from "./mgmt-cluster";

const program = new Command();
program
  .name("ops")
  .description("CLI for performing ops tasks for onmetal.dev")
  .version("0.0.1", "-v, --version", "output the current version");

mgmtCluster(program.command("mgmt-cluster"));

// thread the needle around this bun issue: https://github.com/tj/commander.js/issues/2205
await program.parseAsync(process.argv);
