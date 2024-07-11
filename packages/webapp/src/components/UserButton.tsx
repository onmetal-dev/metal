"use client";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useRouter } from "next/navigation";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import Link from "next/link";
import { useClerk, useUser } from "@clerk/nextjs";

// UserButton is a custom Clerk UserButton component
export const UserButton = () => {
  const { isLoaded, user } = useUser();
  const { signOut, openUserProfile } = useClerk();
  const router = useRouter();
  if (!user?.id) return null;

  return (
    <>
      {!isLoaded ? (
        <Skeleton className="w-8 h-8 rounded-full" />
      ) : (
        <div className="cursor-pointer">
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
              <DropdownMenuGroup>
                <DropdownMenuItem asChild className="cursor-pointer">
                  <Link href="/dashboard">Dashboard</Link>
                </DropdownMenuItem>
                <DropdownMenuItem
                  className="cursor-pointer"
                  onClick={() => openUserProfile()}
                >
                  Profile
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="cursor-pointer"
                onClick={() =>
                  signOut(() => router.push("/"), { redirectUrl: "/" })
                }
              >
                Sign Out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
    </>
  );
};
