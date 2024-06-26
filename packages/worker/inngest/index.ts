import { inngest } from "./client";
import { helloWorld } from "./hello-world";
import { hetznerClusterDelete } from "./hetzner-cluster-delete";
import { hetznerClusterProvision } from "./hetzner-cluster-provision";
import { hetznerProjectCreate } from "./hetzner-project-create";
import { hetznerProjectDelete } from "./hetzner-project-delete";

export { inngest as client };
export const functions = [
  helloWorld,
  hetznerProjectCreate,
  hetznerProjectDelete,
  hetznerClusterProvision,
  hetznerClusterDelete,
];
