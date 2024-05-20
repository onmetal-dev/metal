import { relations, sql } from "drizzle-orm";
import {
  boolean,
  index,
  integer,
  primaryKey,
  text,
  pgEnum,
  timestamp,
  pgSchema,
  varchar,
  json,
  jsonb,
  check,
} from "drizzle-orm/pg-core";
import { customType } from "drizzle-orm/pg-core";
import { createInsertSchema, createSelectSchema } from "drizzle-zod";
import uuidBase62 from "uuid-base62";
import { z } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import sqlSchemaForEnv from "./schemaForEnv";

// use base62 uuids to be shorter and more friendly to the eyes
const uuidId = {
  id: varchar("id", { length: 22 })
    .$defaultFn(() => uuidBase62.v4())
    .primaryKey(),
};

const createdAndUpdatedAt = {
  createdAt: timestamp("created_at")
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
  updatedAt: timestamp("updated_at")
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull()
    .$onUpdateFn(() => new Date()),
};

const schemaDefaults = {
  ...uuidId,
  ...createdAndUpdatedAt,
};

export const metalSchema = pgSchema(
  sqlSchemaForEnv(process.env.NODE_ENV!, process.env.CI)
);

export const users = metalSchema.table(
  "users",
  {
    ...schemaDefaults,
    clerkId: text("clerk_id").unique().notNull(),
    firstName: text("first_name").notNull(),
    lastName: text("last_name").notNull(),
    email: text("email").notNull(),
    emailVerified: boolean("email_verified").notNull().default(false),
    githubId: text("github_id").$type<string | null>(),
  },
  (example) => ({
    clerkIndex: index("clerk_id").on(example.clerkId),
    emailIndex: index("email_idx").on(example.email),
  })
);
export type User = typeof users.$inferSelect; // return type when queried
export type UserInsert = typeof users.$inferInsert;
export const selectUserSchema = createSelectSchema(users);

export const userRelations = relations(users, ({ many }) => ({
  usersToTeams: many(usersToTeams), // user can belong to many teams, a team can have many users
}));

export const teams = metalSchema.table("teams", {
  ...schemaDefaults,
  clerkId: text("clerk_id").unique().notNull(),
  name: text("name").notNull(),
  creatorId: varchar("creator_id", { length: 22 }).notNull(),
});
export type Team = typeof teams.$inferSelect; // return type when queried
export type TeamInsert = typeof teams.$inferInsert;
export const selectTeamSchema = createSelectSchema(teams);

export const teamRelations = relations(teams, ({ one, many }) => ({
  creator: one(users, {
    fields: [teams.creatorId],
    references: [users.id],
  }),
  usersToTeams: many(usersToTeams),
  hetznerProjects: many(hetznerProjects),
  hetznerClusters: many(hetznerClusters),
  applications: many(applications),
  applicationConfigs: many(applicationConfigs),
}));

// usersToTeams tracks a many-to-many relation between users and teams
export const usersToTeams = metalSchema.table(
  "users_to_teams",
  {
    userId: varchar("user_id", { length: 22 })
      .notNull()
      .references(() => users.id),
    teamId: varchar("team_id", { length: 22 })
      .notNull()
      .references(() => teams.id, { onDelete: "cascade", onUpdate: "cascade" }),
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
  ...schemaDefaults,
  creatorId: varchar("creator_id", { length: 22 }).notNull(),
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  hetznerName: text("hetzner_project_name").notNull(),
  hetznerApiToken: text("hetzner_api_token").notNull(),
  sshKeyName: text("ssh_key_name"),
  publicSshKeyData: text("public_ssh_key_data"),
  privateSshKeyData: text("private_ssh_key_data"),
  hetznerWebserviceUsername: text("hetzner_web_service_username").$type<
    string | null
  >(),
  hetznerWebservicePassword: text("hetzner_web_service_password").$type<
    string | null
  >(),
});
export const insertHetznerProjectSchema = createInsertSchema(hetznerProjects);
export type HetznerProjectInsert = z.infer<typeof insertHetznerProjectSchema>;
export const selectHetznerProjectSchema = createSelectSchema(hetznerProjects);
export type HetznerProject = z.infer<typeof selectHetznerProjectSchema>;
export const hetznerProjectSpec = insertHetznerProjectSchema.omit({
  createdAt: true,
  updatedAt: true,
});
export type HetznerProjectSpec = z.infer<typeof hetznerProjectSpec>;

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
export type HetznerNetworkZoneEnum =
  (typeof hetznerNetworkZoneEnum.enumValues)[number];

export const hetznerLocationEnum = pgEnum("hetzner_location_enum", [
  "fsn1", // eu-central
  "nbg1", // eu-central
  "hel1", // eu-central
  "ash", // us-east
  "hil", // us-west
]);
export type HetznerLocationEnum =
  (typeof hetznerLocationEnum.enumValues)[number];

export const hetznerClusterStatusEnum = pgEnum("hetzner_cluster_status_enum", [
  "creating", // waiting for the k8s cluster api to finish its thing
  "initializing", // doing additional installs on the cluster
  "running", // all the nodes are up and running
  "updating", // updating the cluster
  "destroying", // destroying the cluster
  "destroyed", // the cluster is fully destroyed
  "error", // something went wrong
]);
export type HetznerClusterStatusEnum =
  (typeof hetznerClusterStatusEnum.enumValues)[number];

// hetznerClusters. This represents k8s clusters that we spin up within a user's Hetzner account.
export const hetznerClusters = metalSchema.table("hetzner_clusters", {
  ...schemaDefaults,
  creatorId: varchar("creator_id", { length: 22 }).notNull(),
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  hetznerProjectId: varchar("hetzner_project_id", { length: 22 }).notNull(),
  name: text("name").notNull(), // what we call it in the cluster api (assigned)
  status: hetznerClusterStatusEnum("status").notNull(),
  networkZone: hetznerNetworkZoneEnum("network_zone").notNull(),
  location: hetznerLocationEnum("location").notNull(),
  k8sVersion: text("k8s_version").notNull(),
  clusterctlVersion: text("clusterctl_version"),
  clusterctlTemplate: text("clusterctl_manifest"),
  kubeconfig: text("kubeconfig"),
});

export const insertHetznerClusterSchema = createInsertSchema(hetznerClusters);
export type HetznerClusterInsert = z.infer<typeof insertHetznerClusterSchema>;
export const selectHetznerClusterSchema = createSelectSchema(hetznerClusters);
export type HetznerCluster = z.infer<typeof selectHetznerClusterSchema>;

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
  // "cpx11", // 2 vcpu, 2 gb ram, 20 gb disk, 4.35/mo
  // "cpx21", // 3 vcpu, 4 gb ram, 80 gb disk, 7.55/mo
  // "cpx31", // 4 vcpu, 8 gb ram, 160 gb disk, 13.60/mo
  // "cpx41", // 8 vcpu, 16 gb ram, 240 gb disk, 25.20/mo
  // "cpx51", // 16 vcpu, 32 gb ram, 360 gb disk, 54.90/mo
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
export type HetznerInstanceTypeEnum =
  (typeof hetznerInstanceTypeEnum.enumValues)[number];

// hetznerNodeGroups subdivide clusters and provide isolation for different workload types.
export const hetznerNodeGroups = metalSchema.table("hetzner_node_groups", {
  ...schemaDefaults,
  clusterId: varchar("cluster_id", { length: 22 })
    .notNull()
    .references(() => hetznerClusters.id, { onDelete: "cascade" }),
  type: nodeGroupTypeEnum("type").notNull(),
  instanceType: hetznerInstanceTypeEnum("instance_type").notNull(),
  minNodes: integer("min_nodes").notNull(),
  maxNodes: integer("max_nodes").notNull(),
});

export const insertHetznerNodeGroupSchema =
  createInsertSchema(hetznerNodeGroups);
export type HetznerNodeGroupInsert = z.infer<
  typeof insertHetznerNodeGroupSchema
>;
export const selectHetznerNodeGroupSchema =
  createSelectSchema(hetznerNodeGroups);
export const selectHetznerClusterWithNodeGroupsSchema =
  selectHetznerClusterSchema.extend({
    nodeGroups: selectHetznerNodeGroupSchema.array(),
  });
export type HetznerClusterWithNodeGroups = z.infer<
  typeof selectHetznerClusterWithNodeGroupsSchema
>;
export type HetznerNodeGroup = z.infer<typeof selectHetznerNodeGroupSchema>;
export const hetznerNodeGroupSpec = insertHetznerNodeGroupSchema.omit({
  createdAt: true,
  updatedAt: true,
});
export type HetznerNodeGroupSpec = z.infer<typeof hetznerNodeGroupSpec>;
export const hetznerClusterSpec = insertHetznerClusterSchema
  .omit({
    createdAt: true,
    updatedAt: true,
    // fixed for now
    name: true,
    k8sVersion: true,
    // to be filled in when we provision the cluster
    creatorId: true, // implied by request auth
    status: true,
    networkZone: true, // set implicitly by location
    hetznerProjectId: true,
    clusterctlVersion: true,
    clusterctlTemplate: true,
    kubeconfig: true,
  })
  .extend({
    nodeGroups: hetznerNodeGroupSpec.omit({ clusterId: true }).array(),
  });
export type HetznerClusterSpec = z.infer<typeof hetznerClusterSpec>;

export const hetznerNodeGroupRelations = relations(
  hetznerNodeGroups,
  ({ one }) => ({
    hetznerCluster: one(hetznerClusters, {
      fields: [hetznerNodeGroups.clusterId],
      references: [hetznerClusters.id],
    }),
  })
);

const portsSchema = z.array(
  z.object({
    name: z.string(),
    port: z.number(),
    proto: z.enum(["http", "tcp"]),
  })
);
export type Ports = z.infer<typeof portsSchema>;

const externalSchema = z.array(
  z.object({
    name: z.string(),
    portName: z.string(),
    port: z.number().refine((val) => val === 80 || val === 443),
    proto: z.enum(["http", "https"]),
  })
);
export type External = z.infer<typeof externalSchema>;

const envSchema = z.array(z.string());
export type Env = z.infer<typeof envSchema>;

const healthCheckSchema = z.object({
  protocol: z.enum(["http", "tcp"]),
  port: z.union([z.number(), z.string()]),
  path: z.string().optional(),
  httpHeaders: z
    .array(z.object({ name: z.string(), value: z.string() }))
    .optional(),
  initialDelaySeconds: z.number(),
  timeoutSeconds: z.number(),
  periodSeconds: z.number(),
  failureThreshold: z.number(),
  successThreshold: z.number(),
});
export type HealthCheck = z.infer<typeof healthCheckSchema>;

const dependenciesSchema = z.array(z.string());
export type Dependencies = z.infer<typeof dependenciesSchema>;

const databasesSchema = z.array(z.string());
export type Databases = z.infer<typeof databasesSchema>;

const telemetrySchema = z.object({
  traces: z.object({ enabled: z.boolean() }),
  metrics: z.object({ enabled: z.boolean() }),
  logs: z.object({ enabled: z.boolean() }),
});
export type Telemetry = z.infer<typeof telemetrySchema>;

// sourceSchema describes where the source code for an application is located.
// E.g.
// - upload: the team used `metal up` (in cli or ci) to send us a tar.gz that we used to create an application version + build
// - github: the team has connected their github account and at the end of a successful CI run publishes to metal with application version + build info + git metadata. We capture git metadata in source
const sourceSchema = z.object({
  type: z.enum(["upload", "github"]),
  upload: z.object({
    hash: z.string().optional(), // the sha256 of the tar.gz
    path: z.string().optional(), // where we stored it in our object storage
  }),
  github: z.object({
    repository: z.string(), // org/repo
    branch: z.string(),
    commit: z.string(), // full commit sha
  }),
});
export type Source = z.infer<typeof sourceSchema>;

export const applications = metalSchema.table("applications", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  name: text("name").notNull(), // todo: constraint sql`name ~ '^[a-z0-9-]+$'` (when drizzle supports check constraints)
});

const customJsonb = <TData>(name: string) =>
  customType<{ data: TData; driverData: string }>({
    dataType() {
      return "jsonb";
    },
    toDriver(value: TData): string {
      return JSON.stringify(value);
    },
  })(name);

export const applicationConfigs = metalSchema.table(
  "application_configs",
  {
    ...schemaDefaults,
    teamId: varchar("team_id", { length: 22 })
      .notNull()
      .references(() => teams.id, { onDelete: "cascade" }),
    applicationId: varchar("application_id", { length: 22 })
      .notNull()
      .references(() => applications.id, { onDelete: "cascade" }),
    source: customJsonb<Source>("source").notNull(),
    ports: customJsonb<Ports>("ports").notNull(),
    external: customJsonb<External>("external").notNull(),
    env: customJsonb<Env>("env").notNull(),
    healthCheck: customJsonb<HealthCheck>("health_check").notNull(),
    dependencies: customJsonb<Dependencies>("dependencies").notNull(),
    databases: customJsonb<Databases>("databases").notNull(),
    memory: integer("memory").notNull(),
    cpu: integer("cpu").notNull(),
    version: text("version").notNull(),
    // tbd
    // extraStorage: integer("extra_storage").notNull(),
    // autoscaling: json("autoscaling").default(sql`'[]'`),
    // alerts: json("alerts").default(sql`'[]'`),
    // telemetry: customJsonb<Telemetry>("telemetry")
    //   .notNull()
    //   .default(
    //     sql.raw(
    //       `'{}'::jsonb CHECK (jsonb_matches_schema('${telemetryJsonSchema}', telemetry))`
    //     )
    //   ),
  },
  (table) => {
    return {
      versionIdx: index("version_idx").on(table.version),
      createdAtIdx: index("created_at_idx").on(table.createdAt),
    };
  }
);

export const insertApplicationSchema = createInsertSchema(applications);
export type ApplicationInsert = z.infer<typeof insertApplicationSchema>;
export const selectApplicationSchema = createSelectSchema(applications);
export type Application = z.infer<typeof selectApplicationSchema>;

export const insertApplicationConfigSchema =
  createInsertSchema(applicationConfigs);
export type ApplicationConfigInsert = z.infer<
  typeof insertApplicationConfigSchema
>;
export const selectApplicationConfigSchema =
  createSelectSchema(applicationConfigs);
export type ApplicationConfig = z.infer<typeof selectApplicationConfigSchema>;

export const applicationRelations = relations(
  applications,
  ({ one, many }) => ({
    team: one(teams, {
      fields: [applications.teamId],
      references: [teams.id],
    }),
    configs: many(applicationConfigs),
  })
);

export const applicationConfigRelations = relations(
  applicationConfigs,
  ({ one }) => ({
    team: one(teams, {
      fields: [applicationConfigs.teamId],
      references: [teams.id],
    }),
    application: one(applications, {
      fields: [applicationConfigs.applicationId],
      references: [applications.id],
    }),
  })
);
