"use client";
import { joinWaitlist } from "@/app/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ServerActionState, serverActionInitialState } from "@/lib/action";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";
import { useFormState, useFormStatus } from "react-dom";
import confetti from "canvas-confetti";

function Waitlist() {
  const [state, formAction] = useFormState(
    joinWaitlist,
    serverActionInitialState
  );
  const [email, setEmail] = useState("");

  useEffect(() => {
    if (state.message.includes("You've been added to the waitlist!")) {
      confetti();
    }
  }, [state.message]);

  return (
    <form action={formAction}>
      <WaitlistFormContent email={email} setEmail={setEmail} state={state} />
    </form>
  );
}

function WaitlistFormContent({
  email,
  setEmail,
  state,
}: {
  email: string;
  setEmail: (email: string) => void;
  state: ServerActionState;
}) {
  const status = useFormStatus();
  return (
    <>
      <input type="hidden" name="email" value={email} />
      <div className="flex items-center gap-x-4">
        <Input
          // type="email"
          placeholder="you@startup.com"
          className="rounded-3xl px-4 py-2"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={status.pending}
        />
        <Button
          variant="default"
          className="rounded-3xl"
          disabled={status.pending}
        >
          Join Waitlist
        </Button>
      </div>
      {!status.pending && state?.message.length > 0 ? (
        <p
          aria-live="polite"
          className={cn(
            "text-sm pt-4",
            state.isError ? "text-destructive" : "text-success"
          )}
        >
          {state.message}
        </p>
      ) : null}
    </>
  );
}

export default Waitlist;
