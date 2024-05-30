"use client";
import { Skeleton } from "@/components/ui/skeleton";
import { useUser } from "@clerk/nextjs";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export const SignUpButton = () => {
  const { isLoaded, user } = useUser();
  if (!isLoaded) {
    return <Skeleton className="w-16 h-10 rounded-full" />;
  }

  if (!user) {
    return (
      <Button variant="default" className="rounded-3xl" asChild>
        <Link href="/register">Sign up</Link>
      </Button>
    );
  }
  return (
    <Button variant="default" className="rounded-3xl" asChild>
      <Link href="/dashboard">Dashboard</Link>
    </Button>
  );
};
