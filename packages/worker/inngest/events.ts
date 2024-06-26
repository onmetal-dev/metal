type HetznerClusterProvision = {
  data: {
    clusterId: string;
  };
};

type HetznerClusterDelete = {
  data: {
    clusterId: string;
  };
};

import type { HetznerProjectSpec } from "@metal/webapp/app/server/db/schema";

type HetznerProjectCreate = {
  data: HetznerProjectSpec;
};

type HetznerProjectDelete = {
  data: {
    projectId: string;
  };
};

type HelloWorld = {
  data: {
    message: string;
  };
};

export type Events = {
  "hetzner-project/create": HetznerProjectCreate;
  "hetzner-project/delete": HetznerProjectDelete;
  "hetzner-cluster/provision": HetznerClusterProvision;
  "hetzner-cluster/delete": HetznerClusterDelete;
  "test/hello-world": HelloWorld;
};
