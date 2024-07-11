"use client";
import { useCallback, useEffect, useState } from "react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "./ui/command";
import { useCommandStore } from "@/providers/CommandStoreProvider";
import { CommandItem as Item } from "@/stores/CommandStore";
import { useHotkeysContext } from "react-hotkeys-hook";
import _ from "lodash";

function sortByPriority(items: Item[]) {
  return _.orderBy(items, ["priority", "label"], ["desc", "asc"]);
}
export function CommandMenu() {
  const commandState = useCommandStore((state) => state);
  const { enableScope, disableScope, enabledScopes } = useHotkeysContext();
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(!open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, [setOpen]);
  const byGroup = _.groupBy(
    Array.from(commandState.items.values()),
    (item) => item.group ?? "_nogroup_"
  );
  // group order is set by summing priorities within each group
  const groupPriorities = _.mapValues(byGroup, (items) =>
    _.sumBy(items, (item) => item.priority)
  );
  const groups = _.uniq(
    Array.from(commandState.items.values()).map(
      (item) => item.group ?? "_nogroup_"
    )
  );
  const sortedGroups = _.orderBy(groups, (group) => groupPriorities[group], [
    "desc",
  ]);

  // onOpenChange sets state but also enables/disables other hotkeys so that the command menu has full control
  const onOpenChange = useCallback(
    (isOpen: boolean) => {
      setOpen(isOpen);
      if (isOpen) {
        disableScope("*");
      } else {
        enableScope("*");
      }
    },
    [setOpen, enableScope, disableScope]
  );

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        {sortedGroups.map((group) => (
          <CommandGroup
            key={group}
            heading={group === "_nogroup_" ? "" : group}
          >
            {sortByPriority(byGroup[group]!).map((item) => (
              <CommandItem
                key={item.label}
                onSelect={() => {
                  item.onSelect();
                  setOpen(false);
                }}
              >
                <span>{item.label}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}
