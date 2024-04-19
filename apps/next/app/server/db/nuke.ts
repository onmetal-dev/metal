import { db } from "./index";
import { sql } from "drizzle-orm";
import sqlSchemaForEnv from "./schemaForEnv";

const nuke = async (): Promise<void> => {
  const env = process.env.NODE_ENV;
  if (!env || env === "production") {
    console.log("refusing to nuke the db");
    process.exit(1);
  }

  const sqlSchema = sqlSchemaForEnv(process.env.NODE_ENV);
  const tables = await db.execute(
    sql.raw(`SELECT table_name
      FROM information_schema.tables
      WHERE table_schema = '${sqlSchema}'
        AND table_type = 'BASE TABLE';
    `)
  );

  for (let table of tables) {
    const query = sql.raw(`DROP TABLE ${table.table_name} CASCADE;`);
    await db.execute(query);
  }
  console.log("finished nuking the db");
  process.exit(0);
};

nuke();
