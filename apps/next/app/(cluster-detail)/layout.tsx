import { DashboardCommands } from "../dashboard/commands";

export default function ClusterDetailLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen w-full flex">
      <div
        className="w-full h-full scrollbar-none flex px-20 overflow-x-hidden overflow-y-scroll"
        style={{ scrollbarWidth: "none" }}
      >
        <main className="w-full flex-1 items-start max-w-7xl mx-auto">
          <DashboardCommands />
          {children}
        </main>
      </div>
    </div>
  );
}
