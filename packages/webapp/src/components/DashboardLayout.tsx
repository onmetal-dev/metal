"use client";

import { cn } from "@/lib/utils";
import { useStore } from "@/hooks/useStore";
import { Footer } from "@/components/dashboard/Footer";
import { Sidebar } from "@/components/dashboard/Sidebar";
import { useSidebarToggle } from "@/hooks/useSidebarToggle";
import { HotkeysProvider } from "react-hotkeys-hook";
import HotkeysHelpDialog from "./dashboard/HotkeysHelpDialog";
import { CommandMenu } from "./CommandMenu";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const sidebar = useStore(useSidebarToggle, (state) => state);
  if (!sidebar) return null;
  return (
    <>
      <HotkeysProvider initiallyActiveScopes={["*"]}>
        <HotkeysHelpDialog />
        <CommandMenu />
        <Sidebar />
        <main
          className={cn(
            "min-h-[calc(100vh_-_56px)] bg-zinc-50 dark:bg-gray-900 transition-[margin-left] ease-in-out duration-300",
            sidebar?.isOpen === false ? "lg:ml-[90px]" : "lg:ml-72"
          )}
        >
          {children}
        </main>
        <footer
          className={cn(
            "transition-[margin-left] ease-in-out duration-300",
            sidebar?.isOpen === false ? "lg:ml-[90px]" : "lg:ml-72"
          )}
        >
          <Footer />
        </footer>
      </HotkeysProvider>
    </>
  );
}
