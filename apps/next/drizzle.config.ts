import dotenv from "dotenv";
import type { Config } from "drizzle-kit";
import sqlSchemaForEnv from "@/app/server/db/schemaForEnv";
dotenv.config();

let connectionString = process.env.POSTGRES_URL!;

export default {
  schema: "./app/server/db/schema.ts",
  out: "./drizzle",
  driver: "pg",
  dbCredentials: {
    connectionString,
  },
  schemaFilter: sqlSchemaForEnv(process.env.NODE_ENV, process.env.CI),
  strict: process.env.CI !== "true",
  verbose: true,
} satisfies Config;
