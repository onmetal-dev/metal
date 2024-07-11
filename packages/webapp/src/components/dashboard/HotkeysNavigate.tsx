"use client";
import { useRouter } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";

export function HotkeysNavigate({
  hotkeys,
  href,
}: {
  hotkeys?: {
    keys: string | string[];
    description: string;
  };
  href: string;
}) {
  if (!hotkeys) {
    return null;
  }
  const router = useRouter();
  useHotkeys(
    hotkeys.keys,
    () => {
      router.push(href);
    },
    { description: hotkeys.description },
    [router, href]
  );
  return null;
}
