"use client";
import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { useKeyPressEvent } from "react-use";
import { useCommandItems } from "./CommandMenu";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Primitive } from "@radix-ui/react-primitive";

// Context
interface FocusListContextValue<T> {
  data: T[];
  focusedDataIdx: number;
  setFocusedDataIdx: (idx: number) => void;
  focusMode: "mouse" | "keyboard";
  setFocusMode: (mode: "mouse" | "keyboard") => void;
  getHref: (data: T) => string;
}

const FocusListContext = createContext<FocusListContextValue<any> | undefined>(
  undefined
);

function useFocusListContext<T>() {
  const context = useContext(FocusListContext);
  if (!context) {
    throw new Error("FocusList components must be used within a FocusList");
  }
  return context as FocusListContextValue<T>;
}

// Components
export interface FocusListProps<T> {
  data: T[];
  getHref: (data: T) => string;
  children?: React.ReactNode;
  defaultFocusedIdx?: number;
}

export const FocusList = React.forwardRef<HTMLDivElement, FocusListProps<any>>(
  ({ data, getHref, children, defaultFocusedIdx = 0 }, forwardedRef) => {
    const [focusedDataIdx, setFocusedDataIdx] =
      useState<number>(defaultFocusedIdx);
    const [focusMode, setFocusMode] = useState<"mouse" | "keyboard">("mouse");
    const router = useRouter();
    const { open: commandMenuOpen } = useCommandItems();

    const cmdPressed = useRef(false);
    useKeyPressEvent(
      "Meta",
      () => {
        cmdPressed.current = true;
      },
      () => {
        cmdPressed.current = false;
      }
    );

    useKeyPressEvent("j", () => {
      if (commandMenuOpen) return;
      setFocusMode("keyboard");
      if (focusedDataIdx + 1 < data.length)
        setFocusedDataIdx(focusedDataIdx + 1);
    });

    useKeyPressEvent("k", () => {
      if (commandMenuOpen || cmdPressed.current) return;
      setFocusMode("keyboard");
      if (focusedDataIdx > 0) setFocusedDataIdx(focusedDataIdx - 1);
    });

    const cmdMenuLastClosedTime = useRef<null | Date>(null);
    useEffect(() => {
      if (!commandMenuOpen) cmdMenuLastClosedTime.current = new Date();
    }, [commandMenuOpen]);

    useKeyPressEvent("Enter", () => {
      if (
        commandMenuOpen ||
        (cmdMenuLastClosedTime.current &&
          new Date().getTime() - cmdMenuLastClosedTime.current.getTime() < 500)
      )
        return;
      router.push(getHref(data[focusedDataIdx]!));
    });

    const contextValue = {
      data,
      focusedDataIdx,
      setFocusedDataIdx,
      focusMode,
      setFocusMode,
      getHref,
    };

    return (
      <FocusListContext.Provider value={contextValue}>
        <Primitive.div ref={forwardedRef}>{children}</Primitive.div>
      </FocusListContext.Provider>
    );
  }
);

FocusList.displayName = "FocusList";

export const FocusListHeader = React.forwardRef<
  HTMLDivElement,
  FocusListHeaderProps
>(({ children, className }, ref) => (
  <div
    ref={ref}
    className={cn(
      "flex flex-row h-10 px-8 items-center bg-background/60 rounded-t-[7px] shadow-xl text-xs text-muted-foreground",
      "border-b border-muted",
      className
    )}
  >
    {children}
  </div>
));

FocusListHeader.displayName = "FocusListHeader";

export const FocusItems = React.forwardRef<
  HTMLDivElement,
  FocusItemsProps<any>
>(({ children }, ref) => (
  <div ref={ref} className="bg-background rounded-b-[7px] shadow-2xl mb-10">
    {children}
  </div>
));

FocusItems.displayName = "FocusItems";

export const FocusListHead = React.forwardRef<
  HTMLDivElement,
  FocusListHeadProps
>(({ children, className }, ref) => (
  <div ref={ref} className={cn("flex items-center", className)}>
    <h3>{children}</h3>
  </div>
));

FocusListHead.displayName = "FocusListHead";

export const FocusItem = React.forwardRef<HTMLDivElement, FocusItemProps<any>>(
  ({ children, index }, ref) => {
    const {
      data,
      focusedDataIdx,
      focusMode,
      setFocusedDataIdx,
      setFocusMode,
      getHref,
    } = useFocusListContext();

    return (
      <div
        ref={ref}
        onMouseMove={() => {
          if (focusMode !== "mouse") {
            setFocusMode("mouse");
            setFocusedDataIdx(index);
          }
        }}
        onMouseEnter={() => {
          setFocusedDataIdx(index);
        }}
        className={cn(
          "h-11 border-muted rounded-none",
          index !== data.length - 1 && index !== focusedDataIdx - 1
            ? "border-b"
            : "",
          index !== focusedDataIdx ? "text-muted-foreground" : "",
          index === focusedDataIdx
            ? focusMode === "mouse"
              ? "border-2 rounded-sm border-muted-foreground/30"
              : "border-2 rounded-sm border-primary/60"
            : ""
        )}
        style={
          index === focusedDataIdx && focusMode === "keyboard"
            ? { borderStyle: "ridge" }
            : {}
        }
      >
        <Link
          href={getHref(data![index]!)}
          className="text-sm px-8 flex items-center h-full"
        >
          {children}
        </Link>
      </div>
    );
  }
);

FocusItem.displayName = "FocusItem";

export const FocusItemCell = React.forwardRef<
  HTMLDivElement,
  FocusItemCellProps
>(({ children, className }, ref) => (
  <div ref={ref} className={cn("flex items-center", className)}>
    {children}
  </div>
));

FocusItemCell.displayName = "FocusItemCell";

// Types
interface FocusListHeaderProps {
  children?: React.ReactNode;
  className?: string;
}

interface FocusItemsProps<T> {
  children?: React.ReactNode;
}

interface FocusListHeadProps {
  children?: React.ReactNode;
  className?: string;
}

interface FocusItemProps<T> {
  children?: React.ReactNode;
  index: number;
}

interface FocusItemCellProps {
  children?: React.ReactNode;
  className?: string;
}
