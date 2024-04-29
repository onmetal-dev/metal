import Link from "next/link";
import {
  Blocks,
  Database,
  Link2,
  PanelLeft,
  PanelsTopLeft,
  Server,
  SettingsIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { useSelectedLayoutSegment } from "next/navigation";
import { Logo } from "./Logo";

interface NavLinkProps {
  href: string;
  Icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  label: string;
  isActive: boolean;
}

const NavLink = ({ href, Icon, label, isActive }: NavLinkProps) => {
  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-4 px-2.5",
        isActive
          ? "text-foreground"
          : "text-muted-foreground hover:text-foreground"
      )}
    >
      <Icon className="h-5 w-5" />
      {label}
    </Link>
  );
};

export const SidebarNavSm = () => {
  const segment = useSelectedLayoutSegment();

  const links = [
    {
      href: "/dashboard/infrastructure",
      label: "Infrastructure",
      Icon: Server,
    },
    {
      href: "/dashboard/applications",
      label: "Applications",
      Icon: PanelsTopLeft,
    },
    { href: "/dashboard/datastores", label: "Datastores", Icon: Database },
    { href: "/dashboard/add-ons", label: "Add-ons", Icon: Blocks },
    { href: "/dashboard/integrations", label: "Integrations", Icon: Link2 },
    { href: "/dashboard/settings", label: "Settings", Icon: SettingsIcon },
  ];

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button size="icon" variant="outline" className="sm:hidden">
          <PanelLeft className="h-5 w-5" />
          <span className="sr-only">Toggle Menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="sm:max-w-xs">
        <nav className="grid gap-6 text-lg font-medium">
          {/* todo make this the metal logo, similar to sidebarnav */}
          <Link
            href="/dashboard"
            className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground"
          >
            <Logo className="w-9 h-9 transition-all group-hover:scale-110" />
            <span className="sr-only">Acme Inc</span>
          </Link>
          {/* todo: use navlink component */}
          {links.map((link) => (
            <NavLink
              key={link.href}
              href={link.href}
              Icon={link.Icon}
              label={link.label}
              isActive={link.href === segment}
            />
          ))}
        </nav>
      </SheetContent>
    </Sheet>
  );
};
