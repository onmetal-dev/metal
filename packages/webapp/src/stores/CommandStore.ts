import { createStore } from "zustand/vanilla";

export type CommandItem = {
  label: string;
  priority: number;
  group?: string;
  onSelect: () => void;
};

export type CommandState = {
  items: Map<string, CommandItem>;
};

export type CommandActions = {
  addItem: (item: CommandItem) => void;
  removeItem: (label: string) => void;
};

export type CommandStore = CommandState & CommandActions;

export const defaultInitState: CommandState = {
  items: new Map(),
};

export const createCommandStore = (
  initState: CommandState = defaultInitState
) => {
  return createStore<CommandStore>()((set) => ({
    ...initState,
    addItem: (item) =>
      set((state) => {
        if (!state.items.has(item.label)) {
          return { items: new Map(state.items).set(item.label, item) };
        }
        return state;
      }),
    removeItem: (label) =>
      set((state) => {
        if (!state.items.has(label)) {
          return state;
        }
        const newItems = new Map(state.items);
        newItems.delete(label);
        return { items: newItems };
      }),
  }));
};
