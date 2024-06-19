import { SidebarNav } from "@/components/SidebarNav";
import { Topbar } from "@/components/Topbar";
import { DashboardCommands } from "./commands";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen w-full flex-col bg-muted/40">
      <SidebarNav />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14">
        <Topbar />
        <main className="flex-1 items-start p-4 max-w-7xl mx-auto w-full sm:px-6 sm:py-0">
          <DashboardCommands />
          {children}
        </main>
      </div>
    </div>
  );
}
