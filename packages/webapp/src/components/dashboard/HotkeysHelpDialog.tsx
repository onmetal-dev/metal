"use client";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { HotkeysDisplay } from "@/components/ui/keyboard";
import { useState } from "react";
import { useHotkeys, useHotkeysContext } from "react-hotkeys-hook";

export default function HotkeysHelpDialog() {
  const [open, setOpen] = useState(false);
  useHotkeys(
    "shift+?",
    () => {
      setOpen(true);
    },
    { description: "Open keyboard shortcuts help" }
  );
  const { hotkeys: hotkeysRO } = useHotkeysContext();
  const hotkeys = [...hotkeysRO].sort((a, b) =>
    (a.description ?? "").localeCompare(b.description ?? "")
  );
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="text-sm">
        <DialogHeader>
          <DialogTitle>Keyboard Shortcuts</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 text-muted-foreground">
          {hotkeys.map(({ keys, description }, index) => {
            return (
              <div key={index} className="flex justify-between">
                <p>{description}</p>
                <HotkeysDisplay keys={keys ? [...keys] : undefined} />
              </div>
            );
          })}
        </div>
      </DialogContent>
    </Dialog>
  );
}
