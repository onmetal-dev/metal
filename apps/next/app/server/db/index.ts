import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import * as schema from "./schema";

const options = {
  ssl: {
    ca: Buffer.from(process.env.POSTGRES_CA_DATA!, "base64").toString(),
  },
};

export const db = drizzle(postgres(process.env.POSTGRES_URL!, options), {
  schema,
});
