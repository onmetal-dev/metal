import { ModeToggle } from "@/components/ModeToggle";
import { UserNav } from "@/components/dashboard/UserNav";
import { SheetMenu } from "@/components/dashboard/SheetMenu";
import { OrganizationSwitcher } from "@clerk/nextjs";
import { mustGetClerkUser } from "@/app/server/clerk";

interface NavbarProps {
  title: string;
}

export async function Navbar({ title }: NavbarProps) {
  const user = await mustGetClerkUser();
  return (
    <header className="sticky top-0 z-10 w-full bg-background/95 shadow backdrop-blur supports-[backdrop-filter]:bg-background/60 dark:shadow-secondary">
      <div className="mx-4 sm:mx-8 flex h-14 items-center">
        <div className="flex items-center space-x-4 lg:space-x-0">
          <SheetMenu />
          <h1 className="font-bold">{title}</h1>
        </div>
        <div className="flex flex-1 items-center space-x-2 justify-end">
          <OrganizationSwitcher
            hidePersonal
            organizationProfileUrl="/dashboard/settings"
            afterCreateOrganizationUrl="/dashboard"
            afterSelectOrganizationUrl="/dashboard/clusters"
          />
          <ModeToggle />
          <UserNav
            fullName={user.fullName ?? ""}
            initials={
              (user.firstName && user.lastName
                ? `${user.firstName[0]}${user.lastName[0]}`
                : user.fullName
                ? user.fullName[0]
                : "") ?? ""
            }
            email={user.emailAddresses[0]?.emailAddress ?? ""}
            avatarUrl={user.hasImage ? user?.imageUrl : ""}
          />
        </div>
      </div>
    </header>
  );
}
