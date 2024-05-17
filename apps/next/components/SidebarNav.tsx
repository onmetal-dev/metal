"use client";
import { useSelectedLayoutSegment } from "next/navigation";
import Link from "next/link";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "@/components/ui/tooltip";
import { Logo } from "@/components/Logo";
import {
  Server,
  PanelsTopLeft,
  Database,
  Blocks,
  Link2,
  Settings,
} from "lucide-react";

interface NavLinkProps {
  href: string;
  Icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  label: string;
  isActive: boolean;
}

const NavLink = ({ href, Icon, label, isActive }: NavLinkProps) => {
  return (
    <Tooltip key={href}>
      <TooltipTrigger asChild>
        <Link
          href={href}
          className={`flex h-9 w-9 items-center justify-center rounded-lg transition-colors ${
            isActive
              ? "bg-accent text-accent-foreground"
              : "text-muted-foreground hover:text-foreground"
          } md:h-8 md:w-8`}
        >
          <Icon className="h-5 w-5" />
          <span className="sr-only">{label}</span>
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right">{label}</TooltipContent>
    </Tooltip>
  );
};

export const SidebarNav = () => {
  const segment = useSelectedLayoutSegment();
  const isActive = (href: string) => {
    const lastSegment = href.split("/").pop();
    return segment === lastSegment;
  };

  const links = [
    {
      href: "/dashboard/clusters",
      Icon: Server,
      label: "Clusters",
    },
    {
      href: "/dashboard/applications",
      Icon: PanelsTopLeft,
      label: "Applications",
    },
    { href: "/dashboard/datastores", Icon: Database, label: "Datastores" },
    { href: "/dashboard/add-ons", Icon: Blocks, label: "Add-ons" },
    { href: "/dashboard/integrations", Icon: Link2, label: "Integrations" },
  ];

  return (
    <aside className="fixed inset-y-0 left-0 z-10 hidden w-14 flex-col border-r bg-background sm:flex">
      <TooltipProvider>
        <nav className="flex flex-col items-center gap-4 px-2 sm:py-5">
          <Link
            href="/dashboard"
            className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:h-8 md:w-8 md:text-base"
          >
            <Logo className="w-9 h-9 transition-all group-hover:scale-110" />
          </Link>
          {links.map(({ href, Icon, label }) => (
            <NavLink
              key={href}
              href={href}
              Icon={Icon}
              label={label}
              isActive={isActive(href)}
            />
          ))}
        </nav>
        <nav className="mt-auto flex flex-col items-center gap-4 px-2 sm:py-5">
          <NavLink
            key="settings"
            href="/dashboard/settings"
            Icon={Settings}
            label="Settings"
            isActive={isActive("/dashboard/settings")}
          />
        </nav>
      </TooltipProvider>
    </aside>
  );
};
