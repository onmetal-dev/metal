import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { KeySymbol, Keys } from "@/components/ui/keyboard";

export default function Help({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Keyboard Shortcuts</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 text-muted-foreground">
          <div className="flex justify-between">
            <p>Open Help</p>
            <KeySymbol keyName="?" />
          </div>
          <div className="flex justify-between">
            <p>Toggle Time Frame</p>
            <div className="flex flex-row gap-2">
              <KeySymbol keyName="[" />
              <KeySymbol keyName="]" />
            </div>
          </div>
          <div className="flex justify-between">
            <p>Go Back to Clusters Page</p>
            <KeySymbol keyName={Keys.Escape} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
