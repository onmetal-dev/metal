import { sql } from "drizzle-orm";
import {
  boolean,
  index,
  pgTable,
  serial,
  text,
  timestamp,
} from "drizzle-orm/pg-core";

export const users = pgTable(
  "metal_user",
  {
    id: serial("id").primaryKey(),
    clerkId: text("clerk_id").unique().notNull(),
    firstName: text("first_name").notNull(),
    lastName: text("last_name").notNull(),
    email: text("email").notNull(),
    emailVerified: boolean("email_verified").notNull().default(false),
    githubId: text("github_id").$type<string | null>(),
    createdAt: timestamp("created_at")
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
    updatedAt: timestamp("updatedAt"),
  },
  (example) => ({
    clerkIndex: index("clerk_id").on(example.clerkId),
    emailIndex: index("email_idx").on(example.email),
  })
);
export type User = typeof users.$inferSelect; // return type when queried
export type UserInsert = typeof users.$inferInsert;
