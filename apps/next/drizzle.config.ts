import dotenv from "dotenv";
import type { Config } from "drizzle-kit";
import tmp from "tmp";
import { writeFileSync } from "fs";
import sqlSchemaForEnv from "@/app/server/db/schemaForEnv";
dotenv.config();
console.log("process.env.POSTGRES_URL", process.env.POSTGRES_URL);

// Config doesn't have a ca option, so we'll use a tmp file to store the ca
const ca = tmp.fileSync();
writeFileSync(
  ca.name,
  Buffer.from(process.env.POSTGRES_CA_DATA!, "base64").toString()
);

let connectionString = process.env.POSTGRES_URL!;
if (connectionString.includes("?")) {
  connectionString += `&sslrootcert=${ca.name}`;
} else {
  connectionString += `?sslrootcert=${ca.name}`;
}

export default {
  schema: "./app/server/db/schema.ts",
  out: "./drizzle",
  driver: "pg",
  dbCredentials: {
    connectionString,
  },
  schemaFilter: sqlSchemaForEnv(process.env.NODE_ENV),
  strict: process.env.DRIZZLE_DISABLE_STRICT_PUSHES !== "true",
  verbose: true,
} satisfies Config;
