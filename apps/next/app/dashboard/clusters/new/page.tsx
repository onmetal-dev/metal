import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
  CardContent,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Step, StepItem, Stepper } from "@/components/stepper";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { NewClusterForm } from "./form";

export default function NewClusterPage() {
  return (
    <div className="grid flex-1 items-start gap-4 pl-0 pr-4 py-4 sm:py-0 md:gap-8 lg:grid-cols-3">
      <div className="grid auto-rows-max items-start gap-4 lg:gap-8 lg:col-span-2">
        <div className="grid gap-4">
          <NewClusterForm />
        </div>
      </div>
      <div></div>
    </div>
  );
}
