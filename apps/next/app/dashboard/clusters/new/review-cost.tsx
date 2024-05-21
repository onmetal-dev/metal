import { Button } from "@/components/ui/button";
import { useFormStatus } from "react-dom";

interface ReviewAndSubmitProps {
  serverType: {
    prettyPriceMonthly: string; // Assuming it's a string, adjust the type as necessary
  };
  clusterSize: number;
  cost: string;
}
export function ReviewAndSubmit({
  serverType,
  clusterSize,
  cost,
}: ReviewAndSubmitProps) {
  const status = useFormStatus();
  return (
    <>
      <h3 className="text-foreground mb-4 font-bold">
        {`${serverType.prettyPriceMonthly} x ${
          isNaN(clusterSize) ? 0 : clusterSize
        } = ${cost} per month`}
      </h3>
      <Button type="submit" className="mt-2 w-full" disabled={status.pending}>
        Create Cluster
      </Button>
    </>
  );
}
