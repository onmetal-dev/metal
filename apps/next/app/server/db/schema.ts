import { relations, sql } from "drizzle-orm";
import {
  boolean,
  index,
  integer,
  pgTable,
  primaryKey,
  serial,
  text,
  timestamp,
} from "drizzle-orm/pg-core";

export const users = pgTable(
  "metal_users",
  {
    id: serial("id").primaryKey(),
    clerkId: text("clerk_id").unique().notNull(),
    firstName: text("first_name").notNull(),
    lastName: text("last_name").notNull(),
    email: text("email").notNull(),
    emailVerified: boolean("email_verified").notNull().default(false),
    githubId: text("github_id").$type<string | null>(),
    createdAt: timestamp("created_at", { precision: 3 })
      .default(sql`current_timestamp(3)`)
      .notNull(),
    updatedAt: timestamp("updatedAt")
      .default(sql`current_timestamp(3)`)
      .notNull()
      .$onUpdateFn(() => new Date()),
  },
  (example) => ({
    clerkIndex: index("clerk_id").on(example.clerkId),
    emailIndex: index("email_idx").on(example.email),
  })
);
export type User = typeof users.$inferSelect; // return type when queried
export type UserInsert = typeof users.$inferInsert;

export const usersRelations = relations(users, ({ many }) => ({
  usersToTeams: many(usersToTeams), // user can belong to many teams, a team can have many users
}));

export const teams = pgTable("metal_teams", {
  id: serial("id").primaryKey(),
  clerkId: text("clerk_id").unique().notNull(),
  name: text("name").notNull(),
  creatorId: integer("creator_id"),
  createdAt: timestamp("created_at")
    .default(sql`CURRENT_TIMESTAMP`)
    .notNull(),
  updatedAt: timestamp("updatedAt"),
});
export type Team = typeof teams.$inferSelect; // return type when queried
export type TeamInsert = typeof teams.$inferInsert;

export const teamsRelations = relations(teams, ({ one, many }) => ({
  creator: one(users, {
    fields: [teams.creatorId],
    references: [users.id],
  }),
  usersToTeams: many(usersToTeams),
}));

// usersToTeams tracks a many-to-many relation between users and teams
export const usersToTeams = pgTable(
  "metal_users_to_teams",
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
