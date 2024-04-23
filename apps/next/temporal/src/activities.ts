import { ApplicationFailure, Context } from "@temporalio/activity";
import { trace } from "@opentelemetry/api";
import {
  hetznerProjects,
  HetznerProjectSpec,
  HetznerProjectInsert,
  HetznerProject,
} from "@db/schema";
import { db } from "@db/index";
import createClient from "openapi-fetch";
import type { paths } from "@lib/hcloud";
import { generateKeyPairSync, createPublicKey, createHash } from "crypto";
import sshpk from "sshpk";
import { eq } from "drizzle-orm";

export async function createHetznerProject(
  spec: HetznerProjectSpec
): Promise<HetznerProject> {
  return await trace
    .getTracer("metal")
    .startActiveSpan("createHetznerProject", async (span) => {
      span.setAttributes({
        hetznerProjectId: spec.id,
        hetznerName: spec.hetznerName,
        teamId: spec.teamId,
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
            throw ApplicationFailure.create({
              type: "unauthorized",
              message: "API token is invalid",
              nonRetryable: true,
            });
          }
          throw error;
        }
      }

      // use generateKeyPairSync to generate a keypair if none provided
      let publicKey: string = spec.publicSshKeyData ?? "";
      let privateKey: string = spec.privateSshKeyData ?? "";
      if (publicKey === "" && privateKey === "") {
        const keyPair = generateKeyPairSync("ed25519", {
          privateKeyEncoding: { format: "pem", type: "pkcs8" },
          publicKeyEncoding: { format: "pem", type: "spki" },
        });
        publicKey = sshpk.parseKey(keyPair.publicKey, "pem").toString("ssh");
        privateKey = keyPair.privateKey;
      }

      {
        const { error } = await client.POST("/ssh_keys", {
          body: {
            name: `metal-${spec.id}`,
            public_key: publicKey,
          },
        });
        if (error) {
          if ((error as any).error?.code === "uniqueness_error") {
            throw ApplicationFailure.create({
              type: "ssh_key_name_conflict",
              message: `SSH key name metal-${spec.id} already exists, please remove it in Hetzner and try again`,
              nonRetryable: true,
            });
          } else {
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
        publicSshKeyData: Buffer.from(publicKey).toString("base64"),
        privateSshKeyData: Buffer.from(privateKey).toString("base64"),
      };
      const insertResult = await db
        .insert(hetznerProjects)
        .values(insert)
        .returning({ insertedId: hetznerProjects.id });
      if (insertResult.length !== 1) {
        throw new Error("unexpected insert result");
      }

      // select the project
      const selectResult: HetznerProject | undefined = await db
        .select()
        .from(hetznerProjects)
        .where(eq(hetznerProjects.id, insertResult[0]!.insertedId))
        .then((result) => result[0] || undefined);
      if (!selectResult) {
        throw new Error("Project not found");
      }
      return selectResult;
    });
}
