"use client";
import { useRouter, useSelectedLayoutSegment } from "next/navigation";
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
import { Fragment, useEffect } from "react";
import { useKey, useKeyCombo } from "@rwh/react-keystrokes";

interface NavLinkProps {
  href: string;
  Icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  label: string;
  isActive: boolean;
  keyCombo?: string[];
}

const NavLink = ({
  href,
  Icon,
  label,
  isActive,
  keyCombo = [],
}: NavLinkProps) => {
  const router = useRouter();
  const str = keyCombo.map((k) => k.toLowerCase()).join(", ");
  const isKeyComboPressed = useKeyCombo(str);
  useEffect(() => {
    if (isKeyComboPressed) {
      router.push(href);
    }
  }, [isKeyComboPressed, href, router]);
  return (
    <Tooltip key={href}>
      <TooltipTrigger asChild>
        <Link
          href={href}
          className={`flex h-9 w-9 items-center justify-center rounded-sm transition-colors hover:bg-accent hover:text-primary ${
            isActive ? "text-primary" : "text-muted-foreground"
          } md:h-full md:w-full`}
        >
          <Icon className="h-5 w-5" />
          <span className="sr-only">{label}</span>
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right">
        <span>
          {label}
          {keyCombo.length > 0 ? " " : ""}
        </span>
        {keyCombo &&
          keyCombo.map((key, index) => (
            <Fragment key={index}>
              <span className="bg-accent rounded-[3px] p-1 m-1">{key}</span>
              {index < keyCombo.length - 1 && <span> + </span>}
            </Fragment>
          ))}
      </TooltipContent>
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
      keyCombo: ["G", "C"],
    },
    {
      href: "/dashboard/applications",
      Icon: PanelsTopLeft,
      label: "Applications",
      keyCombo: ["G", "A"],
    },
    {
      href: "/dashboard/datastores",
      Icon: Database,
      label: "Datastores",
      keyCombo: ["G", "D"],
    },
    {
      href: "/dashboard/add-ons",
      Icon: Blocks,
      label: "Add-ons",
      keyCombo: ["G", "A"],
    },
    {
      href: "/dashboard/integrations",
      Icon: Link2,
      label: "Integrations",
      keyCombo: ["G", "I"],
    },
  ];

  const activeIndex = links.findIndex((link) => isActive(link.href));

  return (
    <aside className="fixed inset-y-0 left-0 z-10 hidden w-14 flex-col border-r bg-background sm:flex">
      <TooltipProvider>
        <nav className="flex flex-col justify-between items-center gap-4 sm:py-5 h-full">
          <Link
            href="/dashboard"
            className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:h-8 md:w-8 md:text-base"
          >
            <Logo className="w-9 h-9 transition-all group-hover:scale-110" />
          </Link>
          <div className="w-full relative">
            {activeIndex !== -1 && (
              <ActiveIndicator topOffset={48 * activeIndex} />
            )}
            {links.map(({ href, Icon, label, keyCombo }) => (
              <div
                key={href}
                className="h-10 mx-2 mb-2 flex flex-row items-center justify-center"
              >
                <NavLink
                  href={href}
                  Icon={Icon}
                  label={label}
                  isActive={isActive(href)}
                  keyCombo={keyCombo}
                />
              </div>
            ))}
          </div>
          <nav className="flex flex-col items-center gap-4 px-2 w-full relative">
            {isActive("/dashboard/settings") && (
              <ActiveIndicator topOffset={0} />
            )}
            <div className="h-10 mx-2 mb-2 flex flex-row items-center justify-center w-full">
              <NavLink
                key="settings"
                keyCombo={["G", "S"]}
                href="/dashboard/settings"
                Icon={Settings}
                label="Settings"
                isActive={isActive("/dashboard/settings")}
              />
            </div>
          </nav>
        </nav>
      </TooltipProvider>
    </aside>
  );
};

interface ActiveIndicatorProps {
  topOffset: number;
}

const ActiveIndicator: React.FC<ActiveIndicatorProps> = ({ topOffset }) => {
  return (
    <div
      className="bg-primary"
      style={{
        position: "absolute",
        right: "0px",
        top: "0px",
        width: "1px",
        height: "1px",
        margin: "0px auto",
        transformOrigin: "center top",
        willChange: "transform",
        transform: `translateY(${topOffset}px) scaleY(40)`,
        transition: "transform 0.15s ease 0s",
      }}
    />
  );
};
