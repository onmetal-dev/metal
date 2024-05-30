"use client";
import { useSelectedLayoutSegments } from "next/navigation";
import Link from "next/link";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { UserButton } from "@/components/UserButton";
import { SidebarNavSm } from "@/components/SidebarNavSm";
import React from "react";
import { OrganizationSwitcher } from "@clerk/nextjs";

function capFirst(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export const Topbar = () => {
  const segments = useSelectedLayoutSegments();
  // not sure why this doesn't include dashboard...the docs say it should?
  if (segments.length > 0 && segments[0] !== "dashboard") {
    segments.unshift("dashboard");
  }

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
      <SidebarNavSm />
      <Breadcrumb className="hidden md:flex">
        <BreadcrumbList>
          {segments.map((segment, index) => (
            <React.Fragment key={index}>
              <BreadcrumbItem>
                <BreadcrumbLink asChild>
                  <Link href={`/${segments.slice(0, index + 1).join("/")}`}>
                    {capFirst(segment)}
                  </Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              {index < segments.length - 1 && <BreadcrumbSeparator />}
            </React.Fragment>
          ))}
        </BreadcrumbList>
      </Breadcrumb>
      <div className="relative ml-auto flex-1 md:grow-0">
        {/* todo: action menu */}
        {/* <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search..."
              className="w-full rounded-lg bg-background pl-8 md:w-[200px] lg:w-[336px]"
            /> */}
      </div>
      <div className="pt-2">
        <OrganizationSwitcher
          hidePersonal
          organizationProfileUrl="/dashboard/settings"
          createOrganizationUrl="/dashboard/create-team"
        >
          <OrganizationSwitcher.OrganizationProfileLink
            label="Homepage"
            url="/"
            labelIcon={<DotIcon />}
          />
          <OrganizationSwitcher.OrganizationProfilePage
            label="Terms"
            labelIcon={<DotIcon />}
            url="terms"
          >
            <div>
              <h1>Custom Terms Page</h1>
              <p>This is the custom terms page</p>
            </div>
          </OrganizationSwitcher.OrganizationProfilePage>
        </OrganizationSwitcher>
      </div>
      <UserButton />
      {/* todo: I think I want to override the Clerk dropdown to make it feel less Clerk-y */}
      {/* <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="overflow-hidden rounded-full"
              >
                <Image
                  src="/placeholder-user.jpg"
                  width={36}
                  height={36}
                  alt="Avatar"
                  className="overflow-hidden rounded-full"
                />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>My Account</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Settings</DropdownMenuItem>
              <DropdownMenuItem>Support</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Logout</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu> */}
    </header>
  );
};

const DotIcon = () => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 512 512"
      fill="currentColor"
    >
      <path d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512z" />
    </svg>
  );
};
