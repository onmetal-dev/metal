import { relations, sql } from "drizzle-orm";
import {
  boolean,
  customType,
  index,
  integer,
  pgEnum,
  pgSchema,
  primaryKey,
  text,
  timestamp,
  varchar,
} from "drizzle-orm/pg-core";
import { createInsertSchema, createSelectSchema } from "drizzle-zod";
import uuidBase62 from "uuid-base62";
import { z } from "zod";
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
  deployments: many(deployments),
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
  builds: many(builds),
  environments: many(environments),
  sharedVariables: many(sharedVariables),
  appEnvVariables: many(appEnvVariables),
  deployments: many(deployments),
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

const phaseSchema = z.object({
  cmd: z.string().optional(),
  dependsOn: z.array(z.string()).optional(),
});

// builderSchema describes what builder to use (in an application config) or what builder was used (in a build)
const builderSchema = z.object({
  type: z.enum(["nixpacks"]), // todo: "dockerfile", "pack", "image"
  nixpacks: z.object({
    // a subset of what's supported here: https://nixpacks.com/docs/guides/configuring-builds
    providers: z.array(z.string()).optional(),
    buildImage: z.string().optional(),
    phases: z
      .object({
        setup: z.object({
          nixPkgs: z.array(z.string()).optional(),
          nixLibs: z.array(z.string()).optional(),
          aptPkgs: z.array(z.string()).optional(),
        }),
        build: z.object({
          cmd: z.string().optional(),
          dependsOn: z.array(z.string()).optional(),
        }),
        start: z.object({
          cmd: z.string().optional(),
        }),
      })
      .passthrough()
      .refine(
        (data) => {
          // Ensure all additional user-defined phases follow the phase schema
          return Object.entries(data).every(([key, value]) => {
            if (["setup", "build", "start"].includes(key)) {
              return true;
            }
            return phaseSchema.safeParse(value).success;
          });
        },
        {
          message: "All user-defined phases must follow the standard schema",
        }
      ),
  }),
});
export type Builder = z.infer<typeof builderSchema>;

const resourcesSchema = z.object({
  memory: z.number(),
  cpu: z.number(),
});
export type Resources = z.infer<typeof resourcesSchema>;

export const applications = metalSchema.table("applications", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  creatorId: varchar("creator_id", { length: 22 })
    .references(() => users.id)
    .notNull(), // could be null if we can't track down who initiated the deployment
  name: text("name").notNull(), // todo: constraint sql`name ~ '^[a-z0-9-]+$'` (when drizzle supports check constraints)
});

export const applicationRelations = relations(
  applications,
  ({ one, many }) => ({
    team: one(teams, {
      fields: [applications.teamId],
      references: [teams.id],
    }),
    creator: one(users, {
      fields: [applications.creatorId],
      references: [users.id],
    }),
    configs: many(applicationConfigs),
  })
);

export const insertApplicationSchema = createInsertSchema(applications);
export type ApplicationInsert = z.infer<typeof insertApplicationSchema>;
export const selectApplicationSchema = createSelectSchema(applications);
export type Application = z.infer<typeof selectApplicationSchema>;
export const applicationSpec = insertApplicationSchema
  .omit({
    createdAt: true,
    updatedAt: true,
  })
  .refine((data) => /^[a-z0-9-_]+$/.test(data.name), {
    message: "Name must match the regex ^[a-z0-9-_]+$",
    path: ["name"],
  });
export type ApplicationSpec = z.infer<typeof applicationSpec>;

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
    builder: customJsonb<Builder>("builder").notNull(),
    ports: customJsonb<Ports>("ports").notNull(),
    external: customJsonb<External>("external").notNull(),
    env: customJsonb<Env>("env").notNull(),
    healthCheck: customJsonb<HealthCheck>("health_check").notNull(),
    dependencies: customJsonb<Dependencies>("dependencies").notNull(),
    databases: customJsonb<Databases>("databases").notNull(),
    resources: customJsonb<Resources>("resources").notNull(),
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

export const insertApplicationConfigSchema =
  createInsertSchema(applicationConfigs);
export type ApplicationConfigInsert = z.infer<
  typeof insertApplicationConfigSchema
>;
export const selectApplicationConfigSchema =
  createSelectSchema(applicationConfigs);
export type ApplicationConfig = z.infer<typeof selectApplicationConfigSchema>;

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

export const buildStatusEnum = pgEnum("build_status_enum", [
  "pending",
  "running",
  "completed",
  "failed",
]);
export type BuildStatusEnum = (typeof buildStatusEnum.enumValues)[number];

// buildArtifactSchema describes the end result of a build: a built image, or in the case of a serverless app, a tarball with code.
const buildArtifactSchema = z.object({
  image: z
    .object({
      repository: z
        .string()
        .optional()
        .describe(
          "If empty, dockerhub is assumed. Otherwise can be something like registry.k8s.io"
        ),
      name: z
        .string()
        .describe(
          "The name of the image. Required. E.g. busybox, stefanprodan/podinfo."
        ),
      tag: z
        .string()
        .describe("The tag of the image. If not specified, latest is assumed."),
      digest: z
        .string()
        .describe(
          "Instead of a tag you can use a digest. E.g. sha256:1ff6c18fbef2045af6b9c16bf034cc421a29027b800e4f9"
        ),
    })
    .optional(),
  tarball: z
    .object({
      url: z.string(),
    })
    .optional(),
});

export type BuildArtifact = z.infer<typeof buildArtifactSchema>;

export const builds = metalSchema.table("builds", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  applicationId: varchar("application_id", { length: 22 })
    .notNull()
    .references(() => applications.id, { onDelete: "cascade" }),
  applicationConfigId: varchar("application_config_id", { length: 22 })
    .notNull()
    .references(() => applicationConfigs.id, { onDelete: "cascade" }),
  status: buildStatusEnum("status").notNull(),
  logs: text("logs").notNull(),
  artifact: customJsonb<BuildArtifact>("artifact").notNull(),
});

export const buildRelations = relations(builds, ({ one }) => ({
  team: one(teams, {
    fields: [builds.teamId],
    references: [teams.id],
  }),
  application: one(applications, {
    fields: [builds.applicationId],
    references: [applications.id],
  }),
  applicationConfig: one(applicationConfigs, {
    fields: [builds.applicationConfigId],
    references: [applicationConfigs.id],
  }),
}));

export const environments = metalSchema.table("environments", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  name: text("name").notNull(), // todo: constraint sql`name ~ '^[a-z0-9-]+$'` (when drizzle supports check constraints)
});

export const environmentRelations = relations(environments, ({ one }) => ({
  team: one(teams, {
    fields: [environments.teamId],
    references: [teams.id],
  }),
}));

// variablesSchema describes the name and value of an environment variable
const variablesSchema = z.array(
  z.object({
    name: z.string(),
    value: z.string(),
  })
);
export type Variables = z.infer<typeof variablesSchema>;

export const sharedVariables = metalSchema.table("shared_variables", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  variables: customJsonb<Variables>("variables").notNull(),
});

export const sharedVariableRelations = relations(
  sharedVariables,
  ({ one }) => ({
    team: one(teams, {
      fields: [sharedVariables.teamId],
      references: [teams.id],
    }),
  })
);

export const appEnvVariables = metalSchema.table("app_env_variables", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  applicationId: varchar("application_id", { length: 22 })
    .notNull()
    .references(() => applications.id, { onDelete: "cascade" }),
  environmentId: varchar("environment_id", { length: 22 })
    .notNull()
    .references(() => environments.id, { onDelete: "cascade" }),
  variables: customJsonb<Variables>("variables").notNull(),
});

export const appEnvVariableRelations = relations(
  appEnvVariables,
  ({ one }) => ({
    team: one(teams, {
      fields: [appEnvVariables.teamId],
      references: [teams.id],
    }),
    application: one(applications, {
      fields: [appEnvVariables.applicationId],
      references: [applications.id],
    }),
    environment: one(environments, {
      fields: [appEnvVariables.environmentId],
      references: [environments.id],
    }),
  })
);

export const deploymentTypeEnum = pgEnum("deployment_type_enum", [
  "deploy",
  "scale",
  "rollback",
  "restart",
  "suspend",
  "unsuspend",
]);

export const deploymentStatusEnum = pgEnum("deployment_status_enum", [
  "deploying",
  "paused",
  "aborted",
  "resumed",
  "running",
  "cancelled",
  "failed",
  "stopped",
  "stopping",
  "suspended",
]);

export const deployments = metalSchema.table("deployments", {
  ...schemaDefaults,
  teamId: varchar("team_id", { length: 22 })
    .notNull()
    .references(() => teams.id, { onDelete: "cascade" }),
  creatorId: varchar("creator_id", { length: 22 }).references(() => users.id, {
    onDelete: "cascade",
  }), // could be null if we can't track down who initiated the deployment
  applicationId: varchar("application_id", { length: 22 })
    .notNull()
    .references(() => applications.id, { onDelete: "cascade" }),
  applicationConfigId: varchar("application_config_id", { length: 22 })
    .notNull()
    .references(() => applicationConfigs.id, { onDelete: "cascade" }),
  environmentId: varchar("environment_id", { length: 22 })
    .notNull()
    .references(() => environments.id, { onDelete: "cascade" }),
  buildId: varchar("build_id", { length: 22 })
    .notNull()
    .references(() => builds.id, { onDelete: "cascade" }),
  variables: customJsonb<Variables>("variables").notNull(), // snapshot since we want rollbacks to work
  type: deploymentTypeEnum("type").notNull(),
  rolloutStatus: deploymentStatusEnum("rollout_status").notNull(),
  // rolloutStrategy: // within the cluster, the rollout strategy. todo, all-at-once for now
  resources: customJsonb<Resources>("resources").notNull(),
  referenceDeploymentId: varchar("reference_deployment_id", { length: 22 }),
  count: integer("count"), // only set for scale deployments
  // clusterSelector: // todo: clusters you want to deploy into
  // clusterRolloutStrategy: // todo: how to rollout to clusters
  // clusterRolloutStatus: // todo: status of the rollout per cluster
});

export const deploymentRelations = relations(deployments, ({ one }) => ({
  team: one(teams, {
    fields: [deployments.teamId],
    references: [teams.id],
  }),
  user: one(users, {
    fields: [deployments.creatorId],
    references: [users.id],
  }),
  application: one(applications, {
    fields: [deployments.applicationId],
    references: [applications.id],
  }),
  applicationConfig: one(applicationConfigs, {
    fields: [deployments.applicationConfigId],
    references: [applicationConfigs.id],
  }),
  environment: one(environments, {
    fields: [deployments.environmentId],
    references: [environments.id],
  }),
  build: one(builds, {
    fields: [deployments.buildId],
    references: [builds.id],
  }),
}));

// waitlist is a list of users who are on the waitlist
export const waitlistedEmails = metalSchema.table(
  "waitlisted_emails",
  {
    ...schemaDefaults,
    email: varchar("email", { length: 255 }).notNull().unique(),
  },
  (table) => ({
    createdAtIdx: index().on(table.createdAt),
  })
);
export const insertWaitlistedEmailSchema = createInsertSchema(waitlistedEmails);
export type WaitlistedEmailInsert = z.infer<typeof insertWaitlistedEmailSchema>;
export const selectWaitlistedEmailSchema = createSelectSchema(waitlistedEmails);
export type WaitlistedEmail = z.infer<typeof selectWaitlistedEmailSchema>;

// invitedEmails are emails of people who have been invited to join
export const invitedEmails = metalSchema.table("invited_emails", {
  ...schemaDefaults,
  email: varchar("email", { length: 255 }).notNull().unique(),
});

export const insertInvitedEmailSchema = createInsertSchema(invitedEmails);
export type InvitedEmailInsert = z.infer<typeof insertInvitedEmailSchema>;
export const selectInvitedEmailSchema = createSelectSchema(invitedEmails);
export type InvitedEmail = z.infer<typeof selectInvitedEmailSchema>;
