"use client";
import {
  type ReactNode,
  createContext,
  useRef,
  useContext,
  useEffect,
} from "react";
import { useStore } from "zustand";
import {
  type CommandStore,
  createCommandStore,
  CommandItem,
} from "@/stores/CommandStore";

export type CommandStoreApi = ReturnType<typeof createCommandStore>;

export const CommandStoreContext = createContext<CommandStoreApi | undefined>(
  undefined
);

export interface CommandStoreProviderProps {
  children: ReactNode;
}

export const CommandStoreProvider = ({
  children,
}: CommandStoreProviderProps) => {
  const storeRef = useRef<CommandStoreApi>();
  if (!storeRef.current) {
    storeRef.current = createCommandStore();
  }

  return (
    <CommandStoreContext.Provider value={storeRef.current}>
      {children}
    </CommandStoreContext.Provider>
  );
};

export const useCommandStore = <T,>(
  selector: (store: CommandStore) => T
): T => {
  const commandStoreContext = useContext(CommandStoreContext);

  if (!commandStoreContext) {
    throw new Error(`useCommandStore must be used within CommandStoreProvider`);
  }

  return useStore(commandStoreContext, selector);
};

// addCommand does the addItem in a useEffect to avoid setting state in renders and also
// to remove the command on unmount
export function addCommand(item: CommandItem) {
  const { addItem, removeItem } = useCommandStore((state) => state);
  useEffect(() => {
    addItem(item);
    return () => {
      removeItem(item.label);
    };
  }, []);
}
