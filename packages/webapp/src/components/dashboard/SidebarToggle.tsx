"use client";
import { ChevronLeft } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { useHotkeys } from "react-hotkeys-hook";
import { addCommand } from "@/providers/CommandStoreProvider";

interface SidebarToggleProps {
  isOpen: boolean | undefined;
  setIsOpen?: () => void;
}

export function SidebarToggle({ isOpen, setIsOpen }: SidebarToggleProps) {
  addCommand({
    group: "UI",
    label: "Toggle Sidebar",
    priority: 0,
    onSelect: () => {
      setIsOpen?.();
    },
  });
  useHotkeys(
    "shift+<",
    () => {
      if (isOpen) {
        setIsOpen?.();
      }
    },
    { description: "Collapse sidebar" },
    [isOpen, setIsOpen]
  );
  useHotkeys(
    "shift+>",
    () => {
      if (!isOpen) {
        setIsOpen?.();
      }
    },
    { description: "Expand sidebar" },
    [isOpen, setIsOpen]
  );

  return (
    <div className="invisible lg:visible absolute top-[12px] -right-[16px] z-20">
      <Button
        onClick={() => setIsOpen?.()}
        className="w-8 h-8 rounded-md"
        variant="outline"
        size="icon"
      >
        <ChevronLeft
          className={cn(
            "h-4 w-4 transition-transform ease-in-out duration-700",
            isOpen === false ? "rotate-180" : "rotate-0"
          )}
        />
      </Button>
    </div>
  );
}
