"use client";
import { Button } from "@/components/ui/button";
import * as React from "react";
import { Step, type StepItem, Stepper, useStepper } from "@/components/stepper";

function StepperFooterInside({ steps }: { steps: StepItem[] }) {
  return (
    <div className="flex w-full flex-col gap-4">
      <Stepper orientation="vertical" initialStep={0} steps={steps}>
        {steps.map((stepProps, index) => {
          return (
            <Step key={stepProps.label} {...stepProps}>
              <div className="h-40 flex items-center justify-center my-4 border bg-secondary text-primary rounded-md">
                <h1 className="text-xl">Step {index + 1}</h1>
              </div>
              <StepButtons />
            </Step>
          );
        })}
        <FinalStep />
      </Stepper>
    </div>
  );
}

const StepButtons = () => {
  const { nextStep, prevStep, isLastStep, isOptionalStep, isDisabledStep } =
    useStepper();
  return (
    <div className="w-full flex gap-2 mb-4">
      <Button
        disabled={isDisabledStep}
        onClick={prevStep}
        size="sm"
        variant="secondary"
      >
        Prev
      </Button>
      <Button size="sm" onClick={nextStep}>
        {isLastStep ? "Finish" : isOptionalStep ? "Skip" : "Next"}
      </Button>
    </div>
  );
};

const FinalStep = () => {
  const { hasCompletedAllSteps, resetSteps } = useStepper();

  if (!hasCompletedAllSteps) {
    return null;
  }

  return (
    <>
      <div className="h-40 flex items-center justify-center border bg-secondary text-primary rounded-md">
        <h1 className="text-xl">Woohoo! All steps completed! 🎉</h1>
      </div>
      <div className="w-full flex justify-end gap-2">
        <Button size="sm" onClick={resetSteps}>
          Reset
        </Button>
      </div>
    </>
  );
};

export default function Page() {
  const steps = [
    { label: "Log in / Sign up for Hetzner Cloud" },
    { label: "Create a new Hetzner project" },
    { label: "Create a Hetzner API key" },
    { label: "Generate and upload an SSH key" },
  ] satisfies StepItem[];

  return (
    <>
      <StepperFooterInside steps={steps} />
    </>
  );
}
