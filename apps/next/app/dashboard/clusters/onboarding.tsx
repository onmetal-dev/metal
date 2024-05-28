"use client";
import { Button } from "@/components/ui/button";
import * as React from "react";
import { Step, type StepItem, Stepper, useStepper } from "@/components/stepper";
import { Input } from "@/components/ui/input";
import { createHetznerProject } from "./actions";
import { hetznerRedHex, whiteishHex } from "@/lib/constants";
import { useFormState, useFormStatus } from "react-dom";
import { Loader2 } from "lucide-react";
import hetznerLogoImage from "@/images/hetzner-square-200.jpg";
import Image from "next/image";
import { serverActionInitialState, serverActionState } from "./shared";
import { preventDefaultEnter } from "@/lib/utils";

export function Onboarding() {
  const [state, formAction] = useFormState(
    createHetznerProject,
    serverActionInitialState
  );
  const [apiKey, setApiKey] = React.useState("");
  const [projectName, setProjectName] = React.useState("");
  const steps: StepItem[] = [
    { label: "Log in / Sign up for Hetzner Cloud" },
    { label: "Create a new Hetzner project" },
    { label: "Create an API key" },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <div>
          <Image
            src={hetznerLogoImage}
            alt="Hetzner Logo"
            width={50}
            height={50}
          />
        </div>
        <h1 className="text-xl font-semibold">Connect your Hetzner account</h1>
      </div>
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">
          Follow the steps below to connect a Hetzner project and API key to
          Metal.
        </p>
      </div>
      <div className="flex w-full md:max-w-[700px] flex-col gap-4">
        <form action={formAction}>
          <input type="hidden" name="projectName" value={projectName} />
          <input type="hidden" name="apiKey" value={apiKey} />
          <Stepper orientation="vertical" initialStep={0} steps={steps}>
            <Step {...steps[0]}>
              <div className="flex mt-2 mb-4 text-primary rounded-md">
                <HetznerLogin />
              </div>
              <StepButtons />
            </Step>
            <Step {...steps[1]}>
              <div className="flex mt-2 mb-4 text-primary rounded-md">
                <HetznerProject
                  projectName={projectName}
                  setProjectName={setProjectName}
                />
              </div>
              <StepButtons />
            </Step>
            <Step {...steps[2]}>
              <div className="flex mt-2 mb-4 text-primary rounded-md">
                <HetznerApiKey apiKey={apiKey} setApiKey={setApiKey} />
              </div>
              <StepButtons />
              <Result state={state} />
            </Step>
          </Stepper>
        </form>
      </div>
    </div>
  );
}

const Result = ({ state }: { state: serverActionState }) => {
  const status = useFormStatus();
  return (
    <>
      {!status.pending && state?.isError ? (
        <p aria-live="polite" className="text-sm text-destructive">
          {state.message}
        </p>
      ) : null}
    </>
  );
};

const HetznerLogin = () => {
  return (
    <div className="flex flex-col gap-2 ">
      <p className="text-sm text-muted-foreground">
        Return here after successful login
      </p>
      <Button
        asChild
        style={{ backgroundColor: hetznerRedHex, color: whiteishHex }}
      >
        <a href="https://accounts.hetzner.com/login" target="_blank">
          Log in
        </a>
      </Button>
    </div>
  );
};

const HetznerProject = ({
  projectName,
  setProjectName,
}: {
  projectName: string;
  setProjectName: (projectName: string) => void;
}) => {
  const status = useFormStatus();
  return (
    <div className="flex flex-col gap-2 pl-1">
      <p className="text-sm text-muted-foreground">
        Create a new project in the Hetzner Cloud console.
      </p>
      <Button
        asChild
        style={{ backgroundColor: hetznerRedHex, color: whiteishHex }}
        type="button"
      >
        <a href="https://console.hetzner.cloud/projects" target="_blank">
          Create a new project
        </a>
      </Button>
      <p className="text-sm text-muted-foreground">
        Enter the name of the project you created below
      </p>
      <div>
        <Input
          className="text-foreground"
          type="text"
          placeholder="Project name"
          value={projectName}
          onChange={(e) => setProjectName(e.target.value)}
          disabled={status.pending}
          onKeyDown={preventDefaultEnter}
        />
      </div>
    </div>
  );
};

const HetznerApiKey = ({
  apiKey,
  setApiKey,
}: {
  apiKey: string;
  setApiKey: (apiKey: string) => void;
}) => {
  const status = useFormStatus();
  return (
    <div className="flex flex-col gap-2 pl-1">
      <p className="text-sm text-muted-foreground">
        In the project you created, navigate to Security &gt; API tokens &gt;
        Generate API token.
      </p>
      <p className="text-sm text-muted-foreground">Enter the API key below:</p>
      <div>
        <Input
          className="text-foreground"
          type="password"
          placeholder="API key"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          disabled={status.pending}
        />
      </div>
    </div>
  );
};

const StepButtons = () => {
  const { nextStep, prevStep, isLastStep, isOptionalStep, isDisabledStep } =
    useStepper();
  const status = useFormStatus();

  const nextStepFn = (e: React.FormEvent) => {
    e.preventDefault();
    nextStep();
  };

  return (
    <div className="w-full flex gap-2 my-2">
      {!isDisabledStep && (
        <Button onClick={prevStep} size="sm" variant="secondary" type="button">
          Prev
        </Button>
      )}
      {isLastStep ? (
        <>
          <Button size="sm" type="submit" disabled={status.pending}>
            Finish
          </Button>
          {status.pending && (
            <Loader2 className="h-10 w-10 ml-1 text-primary/60 animate-spin" />
          )}
        </>
      ) : (
        <Button size="sm" onClick={nextStepFn} type="button">
          {isOptionalStep ? "Skip" : "Next"}
        </Button>
      )}
    </div>
  );
};
