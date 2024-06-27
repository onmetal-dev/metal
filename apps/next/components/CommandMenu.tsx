"use client";
import {
  useEffect,
  useState,
  ReactNode,
  createContext,
  useContext,
  useCallback,
} from "react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "./ui/command";

type CommandItem = {
  label: string;
  group?: string;
  onSelect: () => void;
};

const CommandItemsContext = createContext<{
  commandItems: Map<string, CommandItem>;
  groupPriorities: Map<string, number>;
  addCommandItem: (item: CommandItem) => void;
  removeCommandItem: (label: string) => void;
  setGroupPriority: (groupName: string, priority: number) => void;
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  commandItems: new Map<string, CommandItem>(),
  groupPriorities: new Map<string, number>(),
  addCommandItem: (item: CommandItem) => {},
  removeCommandItem: (label: string) => {},
  setGroupPriority: (groupName: string, priority: number) => {},
  open: false,
  setOpen: (open: boolean) => {},
});

export function useCommandItems() {
  const context = useContext(CommandItemsContext);
  if (!context) {
    throw new Error(
      "useCommandItems must be used within a CommandItemsProvider"
    );
  }
  return context;
}

export function CommandItemsProvider({ children }: { children: ReactNode }) {
  const [commandItems, setCommandItems] = useState<Map<string, CommandItem>>(
    new Map<string, CommandItem>()
  );
  const [groupPriorities, setGroupPriorities] = useState<Map<string, number>>(
    new Map<string, number>()
  );
  const [open, setOpen] = useState(false);

  const addCommandItem = useCallback((item: CommandItem) => {
    setCommandItems((prevItems) => {
      const newItems = new Map(prevItems);
      newItems.set(item.label, item);
      return newItems;
    });
  }, []);

  const removeCommandItem = useCallback((label: string) => {
    setCommandItems((prevItems) => {
      const newItems = new Map(prevItems);
      newItems.delete(label);
      return newItems;
    });
  }, []);

  const setGroupPriority = useCallback(
    (groupName: string, priority: number) => {
      setGroupPriorities((prevPriorities) => {
        const newPriorities = new Map(prevPriorities);
        newPriorities.set(groupName, priority);
        return newPriorities;
      });
    },
    []
  );

  return (
    <CommandItemsContext.Provider
      value={{
        commandItems,
        groupPriorities,
        addCommandItem,
        removeCommandItem,
        setGroupPriority,
        open,
        setOpen,
      }}
    >
      {children}
    </CommandItemsContext.Provider>
  );
}

export function CommandMenu() {
  // const [open, setOpen] = useState(false);
  const { commandItems, groupPriorities, open, setOpen } = useCommandItems();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(!open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  // Group items by their group property
  const groupedItems: Record<string, CommandItem[]> = Array.from(
    commandItems.values()
  ).reduce<{
    [key: string]: CommandItem[];
  }>((acc, item) => {
    acc[item.group || ""] = acc[item.group || ""] || [];
    acc[item.group || ""]!.push(item);
    return acc;
  }, {});

  // Sort groups by groupPriority
  const sortedGroupKeys: string[] = Object.keys(groupedItems).sort((a, b) => {
    const priorityA = groupPriorities.get(a) || 0;
    const priorityB = groupPriorities.get(b) || 0;
    return priorityB - priorityA;
  });

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        {sortedGroupKeys.length === 0 ? (
          <CommandEmpty>No results found.</CommandEmpty>
        ) : (
          sortedGroupKeys.map((group) => (
            <CommandGroup key={group} heading={group}>
              {groupedItems[group]!.map((item) => (
                <CommandItem
                  key={item.label}
                  onSelect={() => {
                    item.onSelect();
                    setOpen(false);
                  }}
                >
                  {item.label}
                </CommandItem>
              ))}
            </CommandGroup>
          ))
        )}
      </CommandList>
    </CommandDialog>
  );
}
