// adapted from https://stackblitz.com/edit/stackblitz-starters-dlsw2v?file=components%2Fui%2Fkeyboard.tsx

import {
  Fragment,
  HTMLProps,
  ReactNode,
  createContext,
  useContext,
} from "react";
import { cn } from "@/lib/utils";
import { defaultsDeep } from "lodash";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface KeyData {
  symbols: {
    mac?: string;
    windows?: string;
    default: string;
  };
  label: string;
}

export enum Keys {
  Enter = "Enter",
  Space = "Space",
  Control = "Control",
  Shift = "Shift",
  Alt = "Alt",
  Escape = "Escape",
  ArrowUp = "ArrowUp",
  ArrowDown = "ArrowDown",
  ArrowLeft = "ArrowLeft",
  ArrowRight = "ArrowRight",
  Backspace = "Backspace",
  Tab = "Tab",
  CapsLock = "CapsLock",
  Fn = "Fn",
  Command = "Command",
  Insert = "Insert",
  Delete = "Delete",
  Home = "Home",
  End = "End",
  PageUp = "PageUp",
  PageDown = "PageDown",
  PrintScreen = "PrintScreen",
  Pause = "Pause",
}

export const DEFAULT_KEY_MAPPINGS = {
  [Keys.Enter]: {
    symbols: { mac: "↩", default: "↵" },
    label: "Enter",
  },
  [Keys.Space]: {
    symbols: { default: "␣" },
    label: "Space",
  },
  [Keys.Control]: {
    symbols: { mac: "⌃", default: "Ctrl" },
    label: "Control",
  },
  [Keys.Shift]: {
    symbols: { mac: "⇧", default: "Shift" },
    label: "Shift",
  },
  [Keys.Alt]: {
    symbols: { mac: "⌥", default: "Alt" },
    label: "Alt/Option",
  },
  [Keys.Escape]: {
    symbols: { default: "Esc" },
    label: "Escape",
  },
  [Keys.ArrowUp]: {
    symbols: { default: "↑" },
    label: "Arrow Up",
  },
  [Keys.ArrowDown]: {
    symbols: { default: "↓" },
    label: "Arrow Down",
  },
  [Keys.ArrowLeft]: {
    symbols: { default: "←" },
    label: "Arrow Left",
  },
  [Keys.ArrowRight]: {
    symbols: { default: "→" },
    label: "Arrow Right",
  },
  [Keys.Backspace]: {
    symbols: { mac: "⌫", default: "⟵" },
    label: "Backspace",
  },
  [Keys.Tab]: {
    symbols: { mac: "⇥", default: "⭾" },
    label: "Tab",
  },
  [Keys.CapsLock]: {
    symbols: { default: "⇪" },
    label: "Caps Lock",
  },
  [Keys.Fn]: {
    symbols: { default: "Fn" }, // mac symbol for Fn not universally recognized
    label: "Fn",
  },
  [Keys.Command]: {
    symbols: { mac: "⌘", windows: "⊞ Win", default: "Command" },
    label: "Command",
  },
  [Keys.Insert]: {
    symbols: { default: "Ins" },
    label: "Insert",
  },
  [Keys.Delete]: {
    symbols: { mac: "⌦", default: "Del" },
    label: "Delete",
  },
  [Keys.Home]: {
    symbols: { mac: "↖", default: "Home" },
    label: "Home",
  },
  [Keys.End]: {
    symbols: { mac: "↘", default: "End" },
    label: "End",
  },
  [Keys.PageUp]: {
    symbols: { mac: "⇞", default: "PgUp" },
    label: "Page Up",
  },
  [Keys.PageDown]: {
    symbols: { mac: "⇟", default: "PgDn" },
    label: "Page Down",
  },
  [Keys.PrintScreen]: {
    symbols: { default: "PrtSc" },
    label: "Print Screen",
  },
  [Keys.Pause]: {
    symbols: { mac: "⎉", default: "Pause" },
    label: "Pause/Break",
  },
};

interface ShortcutsContextData {
  keyMappings: Record<string, KeyData>;
  os: "mac" | "windows";
}

const ShortcutsContext = createContext<ShortcutsContextData>({
  keyMappings: DEFAULT_KEY_MAPPINGS,
  os: "mac",
});

const useShortcutsContext = () => {
  return useContext(ShortcutsContext);
};

interface ShortcutsProviderProps {
  children: ReactNode;
  keyMappings?: Record<
    string,
    {
      symbols?: {
        mac?: string;
        windows?: string;
        default?: string;
      };
      label?: string;
    }
  >;
  os?: ShortcutsContextData["os"];
}

export const ShortcutsProvider = ({
  children,
  keyMappings = {},
  os = "mac",
}: ShortcutsProviderProps) => {
  const keyMappingsWithDefaults = defaultsDeep(
    {},
    keyMappings,
    DEFAULT_KEY_MAPPINGS
  );
  return (
    <TooltipProvider>
      <ShortcutsContext.Provider
        value={{ keyMappings: keyMappingsWithDefaults, os }}
      >
        {children}
      </ShortcutsContext.Provider>
    </TooltipProvider>
  );
};

interface KeySymbolProps extends HTMLProps<HTMLDivElement> {
  keyName: string;
  disableTooltip?: boolean;
}

export const KeySymbol = ({
  keyName,
  disableTooltip = false,
  className,
  ...otherProps
}: KeySymbolProps) => {
  const context = useShortcutsContext();
  const keyMappings = context.keyMappings;
  const os = context.os || "default";
  const keyData = keyMappings[keyName];
  const symbol = keyData?.symbols?.[os] ?? keyData?.symbols?.default ?? keyName;
  const label = keyData?.label ?? keyName;

  return disableTooltip ? (
    <div
      className={cn(
        "center h-5 min-w-[1.25rem] px-1 w-fit border border-foreground/20 text-foreground/50 text-xs rounded-sm flex flex-row justify-center",
        className
      )}
      {...otherProps}
    >
      <span>{symbol}</span>
    </div>
  ) : (
    <Tooltip delayDuration={300}>
      <TooltipTrigger>
        <div
          className={cn(
            "center h-5 min-w-[1.25rem] px-1 w-fit border border-foreground/20 text-foreground/50 text-xs rounded-sm flex flex-row justify-center",
            className
          )}
          {...otherProps}
        >
          <span>{symbol}</span>
        </div>
      </TooltipTrigger>
      {label !== symbol && (
        <TooltipContent className="px-2 py-1">{label}</TooltipContent>
      )}
    </Tooltip>
  );
};

interface HotkeysDisplayProps {
  keys?: string | string[];
  className?: string;
}

export function HotkeysDisplay({ keys, className }: HotkeysDisplayProps) {
  if (!keys) {
    return null;
  }
  return (
    <div className={cn("flex flex-row", className)}>
      {Array.isArray(keys) ? (
        keys.map((key, index) => (
          <Fragment key={index}>
            <KeySymbol disableTooltip={true} keyName={key} className="ml-2" />
            {index < keys.length - 1 && (
              <span className="ml-1 text-foreground/50 text-xs">+</span>
            )}
          </Fragment>
        ))
      ) : (
        <KeySymbol disableTooltip={true} keyName={keys} className="ml-2" />
      )}
    </div>
  );
}
