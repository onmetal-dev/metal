"use client";

import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import Link from "next/link";
import { KeySymbol } from "@/components/ui/keyboard";
import { HetznerCluster } from "@/app/server/db/schema";

interface TopbarProps {
  cluster: HetznerCluster;
}

interface NavItemProps {
  item: string;
  selected: boolean;
  disabled: boolean;
}

const NavItem = ({ item, selected, disabled }: NavItemProps) => (
  <TooltipProvider>
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          className={`relative px-3 py-2 ${
            disabled
              ? "text-foreground cursor-not-allowed"
              : selected
              ? "text-primary"
              : "text-foreground"
          } group`}
          disabled={disabled}
        >
          {item}
          {selected && (
            <svg
              className="absolute top-0 left-0 right-0 mx-auto"
              width="66%"
              preserveAspectRatio="none"
              height="2"
              viewBox="0 0 20 2"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <line
                x1="0"
                y1="1"
                x2="20"
                y2="1"
                stroke="hsl(221.2 83.2% 53.3%)"
                strokeWidth="2"
              />
            </svg>
          )}
        </button>
      </TooltipTrigger>
      {disabled && (
        <TooltipContent side="top">
          <div>Coming Soon</div>
        </TooltipContent>
      )}
    </Tooltip>
  </TooltipProvider>
);

export const Topbar = ({ cluster }: TopbarProps) => {
  return (
    <TooltipProvider>
      <header
        className="sticky top-0 w-full mx-auto bg-background self-start z-100"
        style={{ backgroundColor: "hsl(var(--background))", zIndex: 1000 }}
      >
        <div className="flex justify-between items-center h-20 max-w-7xl mx-auto">
          {/* lhs header */}
          <div className="flex items-center text-base p-0.5 opacity-50 animate-[0.16s_ease_0s_1_normal_forwards_running] gap-4">
            <Tooltip>
              <TooltipContent side="bottom">
                <div>
                  <span className="mr-2 text-xs">Back</span>
                  <KeySymbol disableTooltip={true} keyName="Escape" />
                </div>
              </TooltipContent>
              <TooltipTrigger asChild>
                <Button variant="secondary" asChild>
                  <Link href="/dashboard/clusters" className="h-8 rounded-sm">
                    <ChevronLeft className="w-4 h-4" />
                  </Link>
                </Button>
              </TooltipTrigger>
            </Tooltip>
            <div>{cluster.name}</div>
          </div>
          {/* rhs header */}
          <div className="flex items-center opacity-50 animate-[0.16s_ease_0s_1_normal_forwards_running]">
            {/* todo */}
            <div className="flex space-x-4">
              {["Metrics", "Alerts"].map((item) => (
                <NavItem
                  key={item}
                  item={item}
                  selected={item === "Metrics"}
                  disabled={item !== "Metrics"}
                />
              ))}
            </div>
          </div>
        </div>
      </header>
    </TooltipProvider>
  );
};
