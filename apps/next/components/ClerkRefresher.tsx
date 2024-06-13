"use client";

import { FC, useEffect, useState } from "react";
import { useClerk, useOrganization } from "@clerk/clerk-react";
import { useVisibilityChange } from "@uidotdev/usehooks";
import { OrganizationResource } from "@clerk/types";
import { useOrganizationList } from "@clerk/nextjs";
import { fetchServerOrgAndSession } from "@/app/dashboard/clusters/actions";

export const ClerkRefresher: FC = () => {
  // console.log("window.Clerk.session", window.Clerk.session);

  const { setActive } = useOrganizationList();
  const documentVisible = useVisibilityChange();
  const clerk = useClerk();
  // const [currentOrg, setCurrentOrg] = useState<OrganizationResource | null>(
  //   null
  // );
  const { organization } = useOrganization();
  // console.log("documentVisible", documentVisible);
  // console.log("organization", organization?.name);
  // console.log("currentOrg", currentOrg?.name);
  // useEffect(() => {
  //   if (documentVisible && organization && organization.id !== currentOrg?.id) {
  //     console.log("Reloading user");
      // clerk.user?.reload();
  //     organization?.reload().then((reloadedOrganization) => {
  //       setActive?.({ organization: reloadedOrganization.id });
  //       console.log("reloadedOrganization", reloadedOrganization);
  //     });
  //     setCurrentOrg(organization);
  //   }
  // }, [organization, documentVisible]);
  useEffect(() => {
    if (!documentVisible) {
      return;
    }
    async function refresh() {
      const { orgId, sessionId } = await fetchServerOrgAndSession();
      console.log("organization.id", organization?.id);
      console.log("orgId", orgId);
      console.log("sessionId", sessionId);
      if (sessionId && orgId && orgId !== organization?.id) {
        await setActive?.({ session: sessionId, organization: orgId });
      }
    }
    refresh();
  }, [documentVisible]);

  return null;
};
