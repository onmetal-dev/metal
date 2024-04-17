import { UserButton as ClerkUserButton } from "@clerk/nextjs";
import { ClerkLoaded, ClerkLoading } from "@clerk/nextjs";
import { Skeleton } from "./ui/skeleton";

// UserButton is a wrapper around the Clerk UserButton component that has a skeleton loader
export const UserButton = () => {
  return (
    <>
      <ClerkLoading>
        <Skeleton className="w-8 h-8 rounded-full" />
      </ClerkLoading>
      <ClerkLoaded>
        <ClerkUserButton afterSignOutUrl="/" />
      </ClerkLoaded>
    </>
  );
};
