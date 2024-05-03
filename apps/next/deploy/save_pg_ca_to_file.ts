import dotenv from "dotenv"
import { writeFileSync } from "fs"

dotenv.config()

if (!process.env.POSTGRES_CA_DATA) {
  throw new Error("POSTGRES_CA_DATA is not set")
}

const ca = Buffer.from(process.env.POSTGRES_CA_DATA, "base64")
writeFileSync("./pg_ca.crt", ca)
