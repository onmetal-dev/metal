import * as React from "react";
import hetznerLogoImage from "@/images/hetzner-square-200.jpg";
import Image from "next/image";
import { Onboarding } from "./onboarding";
import { db } from "@/app/server/db";
import {
  HetznerProject,
  Team,
  hetznerProjects,
  teams,
} from "@/app/server/db/schema";
import { eq } from "drizzle-orm";
import { auth } from "@clerk/nextjs";

const ClusterList = () => {
  return (
    <p>
      Todo: hetzner project connected, now show clusters / create cluster option
    </p>
  );
};

export default async function Page() {
  // load hetzner projects for active org
  const { orgId } = auth();
  if (!orgId) {
    throw new Error("No orgId found");
  }
  const team: Team | undefined = await db
    .select()
    .from(teams)
    .where(eq(teams.clerkId, orgId))
    .then((rows) => rows[0] || undefined);
  if (!team) {
    throw new Error("No team found");
  }
  const hetznerProject: HetznerProject | undefined = await db
    .select()
    .from(hetznerProjects)
    .where(eq(hetznerProjects.teamId, team.id))
    .then((rows) => rows[0] || undefined);
  return (
    <>
      {hetznerProject ? (
        <ClusterList />
      ) : (
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <div>
              <Image
                src={hetznerLogoImage}
                alt="Hetzner Logo"
                width={50}
                height={50}
              />
            </div>
            <h1 className="text-xl font-semibold">
              Connect your Hetzner account
            </h1>
          </div>
          <div className="flex flex-col gap-2">
            <p className="text-sm text-muted-foreground">
              Follow the steps below to connect a Hetzner project and API key to
              Metal.
            </p>
          </div>
          <Onboarding />
        </div>
      )}
    </>
  );
}
