import {
  Blocks,
  Database,
  Link2,
  PanelsTopLeft,
  Server,
  Users,
} from "lucide-react";

export type Submenu = {
  href: string;
  label: string;
  active: boolean;
  hotkeys?: {
    keys: string | string[];
    description: string;
  };
};

export type Menu = {
  href: string;
  label: string;
  hotkeys?: {
    keys: string | string[];
    description: string;
  };
  active: boolean;
  icon: any;
  submenus: Submenu[];
};

export type Group = {
  groupLabel: string;
  menus: Menu[];
};

export function getMenuList(pathname: string): Group[] {
  return [
    {
      groupLabel: "",
      menus: [
        {
          href: "/dashboard/clusters",
          label: "Clusters",
          active: pathname.includes("/dashboard/clusters"),
          icon: Server,
          submenus: [
            {
              href: "/dashboard/clusters",
              label: "All Clusters",
              hotkeys: {
                keys: "c",
                description: "Go to all clusters page",
              },
              active: pathname === "/dashboard/clusters",
            },
            {
              href: "/dashboard/clusters/new",
              label: "New Cluster",
              hotkeys: {
                keys: "v",
                description: "Go to new cluster page",
              },
              active: pathname === "/dashboard/clusters/new",
            },
          ],
        },
      ],
    },
    {
      groupLabel: "",
      menus: [
        {
          href: "/dashboard/applications",
          label: "Applications",
          hotkeys: {
            keys: "a",
            description: "Go to applications page",
          },
          active: pathname.includes("/dashboard/applications"),
          icon: PanelsTopLeft,
          submenus: [],
        },
      ],
    },
    {
      groupLabel: "",
      menus: [
        {
          href: "/dashboard/datastores",
          label: "Datastores",
          hotkeys: {
            keys: "d",
            description: "Go to datastores page",
          },
          active: pathname.includes("/dashboard/datastores"),
          icon: Database,
          submenus: [],
        },
      ],
    },
    {
      groupLabel: "",
      menus: [
        {
          href: "/dashboard/add-ons",
          label: "Add-ons",
          hotkeys: {
            keys: "o",
            description: "Go to add-ons page",
          },
          active: pathname.includes("/dashboard/add-ons"),
          icon: Blocks,
          submenus: [],
        },
      ],
    },
    {
      groupLabel: "",
      menus: [
        {
          href: "/dashboard/integrations",
          label: "Integrations",
          hotkeys: {
            keys: "i",
            description: "Go to integrations page",
          },
          active: pathname.includes("/dashboard/integrations"),
          icon: Link2,
          submenus: [],
        },
      ],
    },
    {
      groupLabel: "Settings",
      menus: [
        {
          href: "/dashboard/settings",
          label: "Team",
          hotkeys: {
            keys: "t",
            description: "Go to team settings page",
          },
          active: pathname.includes("/dashboard/settings"),
          icon: Users,
          submenus: [],
        },
      ],
    },
  ];
}
