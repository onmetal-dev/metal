"use client";
import { UserButton as ClerkUserButton } from "@clerk/nextjs";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useUser, useClerk } from "@clerk/nextjs";
import { useRouter } from "next/navigation";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import Link from "next/link";

interface UserButtonProps
  extends React.ComponentProps<typeof ClerkUserButton> {}

// UserButton is a wrapper around the Clerk UserButton component that has a skeleton loader
export const UserButton = (props: UserButtonProps) => {
  const { isLoaded, user } = useUser();
  const { signOut, openUserProfile } = useClerk();
  const router = useRouter();
  if (!user?.id) return null;

  return (
    <>
      {!isLoaded ? (
        <Skeleton className="w-8 h-8 rounded-full" />
      ) : (
        <>
          {/* <ClerkUserButton {...props} /> */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Avatar>
                <AvatarImage src={user.hasImage ? user?.imageUrl : ""} />
                <AvatarFallback>
                  {`${user?.firstName?.charAt(0)}${user?.lastName?.charAt(0)}`}
                </AvatarFallback>
              </Avatar>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuLabel></DropdownMenuLabel>
              <DropdownMenuGroup>
                <DropdownMenuItem asChild>
                  {/* Create a fictional link to /subscriptions */}
                  <Link href="/dashboard">Dashboard</Link>
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => openUserProfile()}>
                  Profile
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => signOut(() => router.push("/"))}>
                Sign Out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </>
      )}
    </>
  );
};
