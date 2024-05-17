import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import hetznerLocations from "@/lib/hcloud/locations";
import ReactCountryFlag from "react-country-flag";

// for now limit to locations that support ARM
const locationSupportsArm = (locationName: string) => {
  return ["fsn1", "nbg1", "hel1"].includes(locationName);
};

export function ChooseDatacenterStep({
  datacenter,
  setDatacenter,
}: {
  datacenter: string;
  setDatacenter: (value: string) => void;
}) {
  return (
    <TooltipProvider>
      <RadioGroup
        defaultValue={datacenter}
        className="grid grid-cols-3 sm:grid-cols-5 gap-4"
        onValueChange={setDatacenter}
      >
        {hetznerLocations.locations.map((location) => {
          const supportsArm = locationSupportsArm(location.name);
          return (
            <Tooltip key={location.name}>
              <TooltipTrigger asChild disabled={supportsArm}>
                <div>
                  <RadioGroupItem
                    disabled={!supportsArm}
                    value={location.name}
                    id={location.name}
                    className="peer sr-only"
                    aria-label={location.name}
                  />
                  <Label
                    htmlFor={location.name}
                    className="hover:cursor-pointer h-full text-xs flex flex-col items-center justify-center rounded-md border-2 border-muted bg-transparent p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary"
                  >
                    <div className="pb-3">
                      <ReactCountryFlag
                        countryCode={location.country}
                        svg
                        style={{
                          width: "3em",
                          height: "3em",
                        }}
                      />
                    </div>
                    <div className="text-center">{location.city}</div>
                  </Label>
                </div>
              </TooltipTrigger>
              {!supportsArm && (
                <TooltipContent>
                  <span className="text-xs">
                    Currently only deploying ARM servers, which are not
                    supported in this location
                  </span>
                </TooltipContent>
              )}
            </Tooltip>
          );
        })}
      </RadioGroup>
    </TooltipProvider>
  );
}
