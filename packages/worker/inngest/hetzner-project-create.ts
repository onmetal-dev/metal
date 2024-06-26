import type { paths } from "@metal/hcloud";
import { db } from "@metal/webapp/app/server/db";
import type {
  HetznerProject,
  HetznerProjectInsert,
} from "@metal/webapp/app/server/db/schema";
import { hetznerProjects } from "@metal/webapp/app/server/db/schema";
import { tracedExec } from "@metal/webapp/lib/tracedExec";
import { trace } from "@opentelemetry/api";
import { eq } from "drizzle-orm";
import createClient from "openapi-fetch";
import tmp from "tmp";
import { inngest } from "./client";

export const hetznerProjectCreate = inngest.createFunction(
  { id: "hetzner-project-create" },
  { event: "hetzner-project/create" },
  async ({ event, step }) => {
    return await trace
      .getTracer("metal")
      .startActiveSpan("createHetznerProject", async (span) => {
        const spec = event.data;
        span.setAttributes({
          hetznerProjectId: spec.id,
          hetznerName: spec.hetznerName,
          teamId: spec.teamId,
          creatorId: spec.creatorId,
        });
        if (!spec.id) {
          throw new Error("id not provided");
        }
        const client = createClient<paths>({
          headers: { Authorization: `Bearer ${spec.hetznerApiToken}` },
          baseUrl: "https://api.hetzner.cloud/v1",
        });
        {
          const { error } = await client.GET("/datacenters");
          if (error) {
            if ((error as any).error?.code === "unauthorized") {
              return {
                type: "unauthorized",
                message: "API token is invalid",
              };
            }
            throw error;
          }
        }

        let publicKey: string = spec.publicSshKeyData ?? "";
        let privateKey: string = spec.privateSshKeyData ?? "";
        if (publicKey === "" && privateKey === "") {
          const tmpDir = tmp.dirSync();
          await tracedExec({
            spanName: "generate-key",
            spanAttributes: {},
            command: `ssh-keygen -t ed25519 -f ${tmpDir.name}/temp.key -N "" -q`,
          });
          const { stdout: pub } = await tracedExec({
            spanName: "get-public-key",
            spanAttributes: {},
            command: `ssh-keygen -f ${tmpDir.name}/temp.key -y`,
          });
          publicKey = pub;
          let { stdout: pk } = await tracedExec({
            spanName: "get-private-key",
            spanAttributes: {},
            command: `cat ${tmpDir.name}/temp.key`,
          });
          privateKey = pk;
          await tracedExec({
            spanName: "delete-private-key",
            spanAttributes: {},
            command: `rm ${tmpDir.name}/temp.key`,
          });
        }

        const keyName = `metal-${spec.id}`;
        {
          const { data, error: getError } = await client.GET("/ssh_keys");
          if (getError) {
            throw getError;
          }
          const existingKeyWrongPublicKey = data.ssh_keys.find(
            (key: any) => key.name === keyName && key.public_key !== publicKey
          );
          if (existingKeyWrongPublicKey) {
            return {
              type: "ssh_key_name_conflict",
              message: `SSH key name metal-${spec.id} already exists with a different value for the public key, please remove it in Hetzner and try again`,
            };
          }
          const existingKey = data.ssh_keys.find(
            (key: any) => key.name === keyName && key.public_key === publicKey
          );
          if (!existingKey) {
            const { error } = await client.POST("/ssh_keys", {
              body: {
                name: keyName,
                public_key: publicKey,
              },
            });
            if (error) {
              throw error;
            }
          }
        }

        // insert the project
        const insert: HetznerProjectInsert = {
          creatorId: spec.creatorId,
          teamId: spec.teamId,
          hetznerName: spec.hetznerName,
          hetznerApiToken: spec.hetznerApiToken,
          id: spec.id,
          sshKeyName: keyName,
          publicSshKeyData: Buffer.from(publicKey).toString("base64"),
          privateSshKeyData: Buffer.from(privateKey).toString("base64"),
        };
        let insertedId: string | undefined;
        try {
          const insertResult = await db
            .insert(hetznerProjects)
            .values(insert)
            .returning({ insertedId: hetznerProjects.id });
          if (insertResult.length !== 1) {
            throw new Error("unexpected insert result");
          }
          insertedId = insertResult[0]!.insertedId;
        } catch (error: any) {
          if (error.constraint_name === "hetzner_projects_pkey") {
            return {
              type: "hetzner_project_id_conflict",
              message: `Project with id ${spec.id} already exists`,
            };
          } else {
            throw error;
          }
        }

        const selectResult: HetznerProject | undefined = await db
          .select()
          .from(hetznerProjects)
          .where(eq(hetznerProjects.id, insertedId!))
          .then((result) => result[0] || undefined);
        if (!selectResult) {
          throw new Error("Project not found");
        }
        return selectResult;
      });
  }
);
