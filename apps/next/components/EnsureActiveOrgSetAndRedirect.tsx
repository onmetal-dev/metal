"use client";
// inspired by https://clerk.com/docs/guides/force-organizations#set-an-active-organization-based-on-the-url
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth, useOrganizationList } from "@clerk/nextjs";

interface EnsureActiveOrgSetAndRedirectProps {
  activeOrgId: string;
  redirectTo: string;
}
export function EnsureActiveOrgSetAndRedirect({
  activeOrgId,
  redirectTo,
}: EnsureActiveOrgSetAndRedirectProps) {
  const { setActive, isLoaded } = useOrganizationList();
  const router = useRouter();

  // Get the organization ID from the session
  const { orgId } = useAuth();

  useEffect(() => {
    if (!isLoaded) return;

    const setActiveOrgAndRedirect = async () => {
      if (!orgId) {
        await setActive({ organization: activeOrgId });
      }
      router.push(redirectTo);
    };
    setActiveOrgAndRedirect();
  }, [orgId, isLoaded, setActive, activeOrgId, redirectTo, router]);
  return null;
}
