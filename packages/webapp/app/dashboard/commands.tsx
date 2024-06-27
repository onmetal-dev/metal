"use client";
import { useCommandItems } from "@/components/CommandMenu";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export function DashboardCommands() {
  const { addCommandItem, setGroupPriority } = useCommandItems();
  const router = useRouter();
  useEffect(() => {
    setGroupPriority("Pages", 0);
    addCommandItem({
      group: "Pages",
      label: "Clusters",
      onSelect: () => {
        router.push("/dashboard/clusters");
      },
    });
    addCommandItem({
      group: "Pages",
      label: "Applications",
      onSelect: () => {
        router.push("/dashboard/applications");
      },
    });
    addCommandItem({
      group: "Pages",
      label: "Datastores",
      onSelect: () => {
        router.push("/dashboard/datastores");
      },
    });
    addCommandItem({
      group: "Pages",
      label: "Add-ons",
      onSelect: () => {
        router.push("/dashboard/add-ons");
      },
    });
    addCommandItem({
      group: "Pages",
      label: "Integrations",
      onSelect: () => {
        router.push("/dashboard/integrations");
      },
    });
    addCommandItem({
      group: "Pages",
      label: "Settings",
      onSelect: () => {
        router.push("/dashboard/settings");
      },
    });
  }, [addCommandItem, setGroupPriority, router]);
  return null;
}
