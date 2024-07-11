"use client";
import { useRouter } from "next/navigation";
import { addCommand } from "@/providers/CommandStoreProvider";

export function CommandNavigate({
  label,
  priority,
  href,
}: {
  label: string;
  priority?: number;
  href?: string;
}) {
  if (!href) {
    return null;
  }
  const router = useRouter();
  addCommand({
    label,
    group: "Navigate",
    priority: priority ?? 0,
    onSelect: () => {
      router.push(href);
    },
  });
  return null;
}
