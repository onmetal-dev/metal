import { relations, sql } from "drizzle-orm";
import {
  boolean,
  index,
  integer,
  pgTable,
  primaryKey,
  serial,
  text,
  pgEnum,
  timestamp,
  pgSchema,
} from "drizzle-orm/pg-core";
import sqlSchemaForEnv from "./schemaForEnv";

const createdAndUpdatedAt = {
  createdAt: timestamp("created_at")
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
  updatedAt: timestamp("updated_at")
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull()
    .$onUpdateFn(() => new Date()),
};

export const metalSchema = pgSchema(sqlSchemaForEnv(process.env.NODE_ENV));

export const users = metalSchema.table(
  "users",
  {
    id: serial("id").primaryKey(),
    clerkId: text("clerk_id").unique().notNull(),
    firstName: text("first_name").notNull(),
    lastName: text("last_name").notNull(),
    email: text("email").notNull(),
    emailVerified: boolean("email_verified").notNull().default(false),
    githubId: text("github_id").$type<string | null>(),
    ...createdAndUpdatedAt,
  },
  (example) => ({
    clerkIndex: index("clerk_id").on(example.clerkId),
    emailIndex: index("email_idx").on(example.email),
  })
);
export type User = typeof users.$inferSelect; // return type when queried
export type UserInsert = typeof users.$inferInsert;

export const userRelations = relations(users, ({ many }) => ({
  usersToTeams: many(usersToTeams), // user can belong to many teams, a team can have many users
}));

export const teams = metalSchema.table("teams", {
  id: serial("id").primaryKey(),
  clerkId: text("clerk_id").unique().notNull(),
  name: text("name").notNull(),
  creatorId: integer("creator_id"),
  ...createdAndUpdatedAt,
});
export type Team = typeof teams.$inferSelect; // return type when queried
export type TeamInsert = typeof teams.$inferInsert;

export const teamRelations = relations(teams, ({ one, many }) => ({
  creator: one(users, {
    fields: [teams.creatorId],
    references: [users.id],
  }),
  usersToTeams: many(usersToTeams),
  hetznerProjects: many(hetznerProjects),
  hetznerClusters: many(hetznerClusters),
}));

// usersToTeams tracks a many-to-many relation between users and teams
export const usersToTeams = metalSchema.table(
  "users_to_teams",
  {
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    teamId: integer("team_id")
      .notNull()
      .references(() => teams.id),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.userId, t.teamId] }),
  })
);
export const usersToTeamsRelations = relations(usersToTeams, ({ one }) => ({
  team: one(teams, {
    fields: [usersToTeams.teamId],
    references: [teams.id],
  }),
  user: one(users, {
    fields: [usersToTeams.userId],
    references: [users.id],
  }),
}));

// hetznerProjects. Hetzner requires all resources to be created within the confines of a project.
// A Hetzner project is also what we ask users to connect to Metal.
// This table contains the information provided by users about a Hetzner project in their Hetzner account.
// Once we have this Hetzner project info we can start creating things in their Hetzner account.
// We associate them with the user who created them, the team they belong to, and the clusters they contain.
export const hetznerProjects = metalSchema.table("hetzner_projects", {
  id: serial("id").primaryKey(),
  creatorId: integer("creator_id").notNull(),
  teamId: integer("team_id").notNull(),
  hetznerName: text("hetzner_project_name").notNull(),
  hetznerApiToken: text("hetzner_api_token").notNull(),
  publicSshKeyData: text("public_ssh_key_data").notNull(),
  privateSshKeyData: text("private_ssh_key_data").notNull(),
  hetznerWebserviceUsername: text("hetzner_web_service_username").$type<
    string | null
  >(),
  hetznerWebservicePassword: text("hetzner_web_service_password").$type<
    string | null
  >(),
  ...createdAndUpdatedAt,
});

export const hetznerProjectRelations = relations(
  hetznerProjects,
  ({ one, many }) => ({
    creatorId: one(users, {
      fields: [hetznerProjects.creatorId],
      references: [users.id],
    }),
    teamId: one(teams, {
      fields: [hetznerProjects.teamId],
      references: [teams.id],
    }),
    clusters: many(hetznerClusters),
  })
);

// Hetzner thinks of geography as "network zones" at the highest level.
// Within each network zone are locations.
export const hetznerNetworkZoneEnum = pgEnum("hetzner_network_zone_enum", [
  "eu-central",
  "us-east",
  "us-west",
]);
export const hetznerLocationEnum = pgEnum("hetzner_location_enum", [
  "fsn1", // eu-central
  "nbg1", // eu-central
  "hel1", // eu-central
  "ash", // us-east
  "hil", // us-west
]);

export const hetznerClusterStatusEnum = pgEnum("hetzner_cluster_status_enum", [
  "creating", // waiting for the k8s cluster api to finish its thing
  "initializing", // doing additional installs on the cluster
  "running", // all the nodes are up and running
  "updating", // updating the cluster
  "destroying", // destroying the cluster
  "destroyed", // the cluster is fully destroyed
  "error", // something went wrong
]);

// hetznerClusters. This represents k8s clusters that we spin up within a user's Hetzner account.
export const hetznerClusters = metalSchema.table("hetzner_clusters", {
  id: serial("id").primaryKey(),
  creatorId: integer("creator_id").notNull(),
  teamId: integer("team_id").notNull(),
  hetznerProjectId: integer("hetzner_project_id").notNull(),
  name: text("name").notNull(), // what we call it in the cluster api (assigned)
  vanityName: text("vanity_name").notNull(), // name the user gave it
  status: hetznerClusterStatusEnum("status").notNull(),
  networkZone: hetznerNetworkZoneEnum("network_zone").notNull(),
  cidr: text("cidr").notNull(),
  location: hetznerLocationEnum("location").notNull(),
  k8sVersion: text("k8s_version").notNull(),
  ...createdAndUpdatedAt,
});

export const hetznerClusterRelations = relations(
  hetznerClusters,
  ({ one, many }) => ({
    creatorId: one(users, {
      fields: [hetznerClusters.creatorId],
      references: [users.id],
    }),
    teamId: one(teams, {
      fields: [hetznerClusters.creatorId],
      references: [teams.id],
    }),
    hetznerProjectId: one(hetznerProjects, {
      fields: [hetznerClusters.creatorId],
      references: [hetznerProjects.id],
    }),
    nodeGroups: many(hetznerNodeGroups),
  })
);

export const nodeGroupTypeEnum = pgEnum("node_group_type_enum", ["all"]); // future support for different types as a cluster gets bigger: "application", "system", "monitoring", "ingress"
export const hetznerInstanceTypeEnum = pgEnum("hetzner_instance_type_enum", [
  // Cloud Shared AMD and Cloud Dedicated AMD are available in all locations, so list those out
  // https://docs.hetzner.com/cloud/general/locations/#which-cloud-products-are-available
  // amd shared: (knock off 0.50 for ipv6 only. all come with 20 TB traffic)
  "cpx11", // 2 vcpu, 2 gb ram, 20 gb disk, 4.35/mo
  "cpx21", // 3 vcpu, 4 gb ram, 80 gb disk, 7.55/mo
  "cpx31", // 4 vcpu, 8 gb ram, 160 gb disk, 13.60/mo
  "cpx41", // 8 vcpu, 16 gb ram, 240 gb disk, 25.20/mo
  "cpx51", // 16 vcpu, 32 gb ram, 360 gb disk, 54.90/mo
  // amd dedicated: (also knock off 0.50 for ipv6 only. range from 20 to 60 TB traffic)
  // "ccx13", // 2 vcpu, 8 gb ram, 80 gb disk, 12.49/mo
  // "ccx23", // 4 vcpu, 16 gb ram, 160 gb disk, 24.49/mo
  // "ccx33", // 8 vcpu, 32 gb ram, 240 gb disk, 48.49/mo
  // "ccx43", // 16 vcpu, 64 gb ram, 360 gb disk, 96.49/mo
  // "ccx53", // 32 vcpu, 128 gb ram, 600 gb disk, 192.49/mo
  // "ccx63", // 48 vcpu, 192 gb ram, 950 gb disk, 288.49/mo
  // Cloud Shared ARM is cheap, but
  // - there would probably be difficulties building apps targeting ARM
  // - only available in fsn1, nbg1, hel1 (no us)
  // (knock off 0.50 for ipv6 only. all come with 20 TB traffic)
  "cax11", // 2 vcpu, 4 gb ram, 40 gb disk, 3.79/mo
  "cax21", // 4 vcpu, 8 gb ram, 80 gb disk, 6.49/mo
  "cax31", // 8 vcpu, 16 gb ram, 160 gb disk, 12.49/mo
  "cax41", // 16 vcpu, 32 gb ram, 320 gb disk, 24.49/mo
]);

// hetznerNodeGroups subdivide clusters and provide isolation for different workload types.
export const hetznerNodeGroups = metalSchema.table("hetzner_node_group", {
  id: serial("id").primaryKey(),
  clusterId: integer("cluster_id").notNull(),
  type: nodeGroupTypeEnum("type").notNull(),
  instanceType: hetznerInstanceTypeEnum("instance_type").notNull(),
  minNodes: integer("min_nodes").notNull(),
  maxNodes: integer("max_nodes").notNull(),
  ...createdAndUpdatedAt,
});

export const hetznerNodeGroupRelations = relations(
  hetznerNodeGroups,
  ({ one }) => ({
    hetznerCluster: one(hetznerClusters, {
      fields: [hetznerNodeGroups.clusterId],
      references: [hetznerClusters.id],
    }),
  })
);
