import sqlSchemaForEnv from "@/app/server/db/schemaForEnv";
import dotenv from "dotenv";
import type { Config } from "drizzle-kit";
dotenv.config();

export default {
  schema: "./app/server/db/schema.ts",
  out: "./drizzle",
  dialect: "postgresql",
  dbCredentials: {
    url: process.env.POSTGRES_URL!,
  },
  schemaFilter: sqlSchemaForEnv(process.env.NODE_ENV, process.env.CI),
  strict: process.env.CI !== "true",
  verbose: true,
  migrations: {
    table: "migrations",
    schema: sqlSchemaForEnv(process.env.NODE_ENV, process.env.CI),
  },
} satisfies Config;
