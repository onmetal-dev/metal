"use client";

import { useRouter } from "next/navigation";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { CreateOrganization } from "@clerk/nextjs";

type PointerDownOutsideEvent = CustomEvent<{
  originalEvent: PointerEvent;
}>;

export default function CreateTeamPage() {
  const router = useRouter();
  const onDismiss = (event: PointerDownOutsideEvent | KeyboardEvent) => {
    event.preventDefault();
    event.stopPropagation();
    router.replace("/dashboard");
  };

  return (
    <Dialog defaultOpen>
      <DialogContent
        onPointerDownOutside={onDismiss}
        onEscapeKeyDown={onDismiss}
      >
        <CreateOrganization afterCreateOrganizationUrl="/dashboard" />
      </DialogContent>
    </Dialog>
  );
}
