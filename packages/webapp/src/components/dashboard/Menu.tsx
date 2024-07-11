"use client";
import Link from "next/link";
import { Ellipsis, LogOut } from "lucide-react";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { getMenuList } from "@/lib/getMenuList";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { CollapseMenuButton } from "@/components/dashboard/CollapseMenuButton";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "@/components/ui/tooltip";
import { Fragment } from "react";
import { HotkeysDisplay } from "../ui/keyboard";
import { HotkeysNavigate } from "./HotkeysNavigate";
import { CommandNavigate } from "./CommandNavigate";
import { RemoveScrollBar } from "react-remove-scroll-bar";

interface MenuProps {
  isOpen: boolean | undefined;
}

export function Menu({ isOpen }: MenuProps) {
  const pathname = usePathname();
  const menuList = getMenuList(pathname);
  return (
    <div>
      <RemoveScrollBar />
      <ScrollArea className="[&>div>div[style]]:!block" scrollHideDelay={1}>
        <>
          {menuList.map(({ menus }, index1) =>
            menus.map(({ href, submenus, hotkeys, label }, index2) => (
              <Fragment key={index1 * 100 + index2}>
                <HotkeysNavigate hotkeys={hotkeys} href={href} />
                {submenus.length === 0 && (
                  <CommandNavigate label={label} priority={0} href={href} />
                )}
                {submenus.map(({ href, hotkeys, label }, index3) => (
                  <Fragment key={index1 * 100 + index2 * 100 + index3}>
                    <HotkeysNavigate hotkeys={hotkeys} href={href} />
                    <CommandNavigate label={label} priority={0} href={href} />
                  </Fragment>
                ))}
              </Fragment>
            ))
          )}
        </>
        <nav className="w-full h-full mt-8">
          <ul className="flex flex-col min-h-[calc(100vh-48px-36px-16px-32px)] lg:min-h-[calc(100vh-32px-40px-32px)] items-start space-y-1 px-2">
            {menuList.map(({ groupLabel, menus }, index) => (
              <li
                className={cn("w-full", groupLabel ? "pt-5" : "")}
                key={index}
              >
                {(isOpen && groupLabel) || isOpen === undefined ? (
                  <p className="text-sm font-medium text-muted-foreground px-4 pb-2 max-w-[248px] truncate">
                    {groupLabel}
                  </p>
                ) : !isOpen && isOpen !== undefined && groupLabel ? (
                  <TooltipProvider>
                    <Tooltip delayDuration={100}>
                      <TooltipTrigger className="w-full">
                        <div className="flex items-center justify-center w-full">
                          <Ellipsis className="w-5 h-5" />
                        </div>
                      </TooltipTrigger>
                      <TooltipContent side="right">
                        <p>{groupLabel}</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                ) : (
                  <p className="pb-2"></p>
                )}
                {menus.map(
                  (
                    { href, label, icon: Icon, active, submenus, hotkeys },
                    index
                  ) =>
                    submenus.length === 0 ? (
                      <div className="w-full" key={index}>
                        <TooltipProvider disableHoverableContent>
                          <Tooltip delayDuration={100}>
                            <TooltipTrigger asChild>
                              <Button
                                variant={active ? "secondary" : "ghost"}
                                className="justify-between w-full h-10 mb-1 group"
                                asChild
                              >
                                <Link
                                  href={href}
                                  className="flex items-center justify-between w-full"
                                >
                                  <div className="flex items-center">
                                    <span
                                      className={cn(
                                        isOpen === false ? "" : "mr-4"
                                      )}
                                    >
                                      <Icon size={18} />
                                    </span>
                                    <p
                                      className={cn(
                                        "max-w-[200px] truncate",
                                        isOpen === false
                                          ? "-translate-x-96 opacity-0"
                                          : "translate-x-0 opacity-100"
                                      )}
                                    >
                                      {label}
                                    </p>
                                  </div>
                                  <HotkeysDisplay
                                    keys={hotkeys?.keys}
                                    className="transition-opacity duration-300 opacity-0 group-hover:opacity-100"
                                  />
                                </Link>
                              </Button>
                            </TooltipTrigger>
                            {isOpen === false && (
                              <TooltipContent
                                side="right"
                                className="flex flex-row items-center"
                              >
                                <span>{label}</span>
                                <HotkeysDisplay keys={hotkeys?.keys} />
                              </TooltipContent>
                            )}
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                    ) : (
                      <div className="w-full" key={index}>
                        <CollapseMenuButton
                          icon={Icon}
                          label={label}
                          active={active}
                          submenus={submenus}
                          isOpen={isOpen}
                        />
                      </div>
                    )
                )}
              </li>
            ))}
            <li className="flex items-end w-full grow">
              <TooltipProvider disableHoverableContent>
                <Tooltip delayDuration={100}>
                  <TooltipTrigger asChild>
                    <Button
                      onClick={() => {}}
                      variant="outline"
                      className="justify-center w-full h-10 mt-5"
                    >
                      <span className={cn(isOpen === false ? "" : "mr-4")}>
                        <LogOut size={18} />
                      </span>
                      <p
                        className={cn(
                          "whitespace-nowrap",
                          isOpen === false ? "opacity-0 hidden" : "opacity-100"
                        )}
                      >
                        Sign out
                      </p>
                    </Button>
                  </TooltipTrigger>
                  {isOpen === false && (
                    <TooltipContent side="right">Sign out</TooltipContent>
                  )}
                </Tooltip>
              </TooltipProvider>
            </li>
          </ul>
        </nav>
      </ScrollArea>
    </div>
  );
}
