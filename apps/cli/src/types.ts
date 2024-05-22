import { type WhoAmI } from "@onmetal/node/resources/whoami.mjs";

type AppLink = {
  directory: string;
  appId: string;
  appName: string;
  teamId: string;
  teamName: string;
};

export interface Config {
  whoami?: WhoAmI;
  appLinks?: {
    [directory: string]: AppLink;
  };
}
