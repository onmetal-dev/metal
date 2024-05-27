"use client";

import { Fragment } from "react";
import Link from "next/link";
import { Popover, Transition } from "@headlessui/react";
import clsx from "clsx";

import { Button } from "@/components/ui/button";
import { Container } from "@/components/Container";
import { LogoWide } from "@/components/LogoWide";
import { NavLink } from "@/components/NavLink";
import { Menu, X } from "lucide-react";
import { SignInButton } from "@clerk/nextjs";
import { SignedIn, SignedOut } from "@clerk/clerk-react";
import { useTheme } from "next-themes";
import { UserButton } from "./UserButton";

function MobileNavLink({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) {
  return (
    <Popover.Button as={Link} href={href} className="block w-full p-2">
      {children}
    </Popover.Button>
  );
}

function MobileNavIcon({ open }: { open: boolean }) {
  return (
    <>
      <Transition
        as={Fragment}
        show={open}
        enter="transform transition duration-[400ms]"
        enterFrom="opacity-0 rotate-[-120deg] scale-50"
        enterTo="opacity-100 rotate-0 scale-100"
        // leave immediately to make room for the other icon
        // leave="transform duration-200 transition ease-in-out"
        leaveFrom="opacity-100 rotate-0 scale-100 "
        leaveTo="opacity-0 scale-95 "
      >
        <X className={clsx("h-5.5 w-5.5 stroke-foreground")} />
      </Transition>
      <Transition
        as={Fragment}
        show={!open}
        enter="transform transition duration-[400ms]"
        enterFrom="opacity-0 rotate-[-120deg] scale-50"
        enterTo="opacity-100 rotate-0 scale-100"
        // leave immediately to make room for the other icon
        // leave="transform duration-200 transition ease-in-out"
        leaveFrom="opacity-100 rotate-0 scale-100 "
        leaveTo="opacity-0 scale-95 "
      >
        <Menu className={clsx("h-5.5 w-5.5 stroke-foreground")} />
      </Transition>
    </>
  );
}

function MobileNavigation() {
  return (
    <Popover>
      <Popover.Button
        className="relative z-10 flex items-center justify-center w-8 h-8 ui-not-focus-visible:outline-none"
        aria-label="Toggle Navigation"
      >
        {({ open }) => <MobileNavIcon open={open} />}
      </Popover.Button>
      <Transition.Root>
        <Transition.Child
          as={Fragment}
          enter="duration-150 ease-out"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="duration-150 ease-in"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <Popover.Overlay className="fixed inset-0 bg-muted/70" />
        </Transition.Child>
        <Transition.Child
          as={Fragment}
          enter="duration-150 ease-out"
          enterFrom="opacity-0 scale-95"
          enterTo="opacity-100 scale-100"
          leave="duration-100 ease-in"
          leaveFrom="opacity-100 scale-100"
          leaveTo="opacity-0 scale-95"
        >
          <Popover.Panel
            as="div"
            className="absolute inset-x-0 flex flex-col p-4 mt-4 text-lg tracking-tight origin-top bg-background shadow-xl top-full rounded-2xl text-foreground"
          >
            <MobileNavLink href="#features">Features</MobileNavLink>
            {/* <MobileNavLink href="#testimonials">Testimonials</MobileNavLink> */}
            <MobileNavLink href="#pricing">Pricing</MobileNavLink>
            <hr className="m-2 border-border" />
            <MobileNavLink href="/login">Sign in</MobileNavLink>
          </Popover.Panel>
        </Transition.Child>
      </Transition.Root>
    </Popover>
  );
}

export function Header() {
  const { theme } = useTheme();
  return (
    <header className="py-10">
      <Container>
        <nav className="relative z-50 flex justify-between">
          <div className="flex items-center md:gap-x-12">
            <Link href="#" aria-label="Home">
              <LogoWide className="w-auto h-10" />
            </Link>
            <div className="hidden md:flex md:gap-x-6">
              <NavLink href="#features">Features</NavLink>
              {/* <NavLink href="#testimonials">Testimonials</NavLink> */}
              <NavLink href="#pricing">Pricing</NavLink>
            </div>
          </div>
          <div className="flex items-center gap-x-5 md:gap-x-8">
            <SignedOut>
              <div className="hidden md:block">
                <Button
                  variant="outline"
                  className="rounded-3xl cursor-pointer"
                  asChild
                >
                  <SignInButton mode="modal">
                    <span>Sign in</span>
                  </SignInButton>
                </Button>
              </div>
              <Button variant="default" className="rounded-3xl" asChild>
                <Link href="/register">Get started today</Link>
              </Button>
            </SignedOut>
            <SignedIn>
              <UserButton
                afterSignOutUrl="/"
                appearance={{
                  variables: {
                    colorBackground: "hsl(217.2 32.6% 17.5%)", // --muted
                  },
                }}
              />
            </SignedIn>
            <div className="-mr-1 md:hidden">
              <MobileNavigation />
            </div>
          </div>
        </nav>
      </Container>
    </header>
  );
}
