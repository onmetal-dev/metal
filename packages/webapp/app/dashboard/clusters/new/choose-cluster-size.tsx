import { Input } from "@/components/ui/input";
import { preventDefaultEnter } from "@/lib/utils";

interface ChooseClusterSizeProps {
  clusterSize: number;
  setClusterSize: (size: number) => void;
}

export function ChooseClusterSize({
  clusterSize,
  setClusterSize,
}: ChooseClusterSizeProps) {
  return (
    <Input
      className="mb-4 text-foreground"
      value={clusterSize}
      type="number"
      id="cluster-size"
      min="1"
      max="100"
      onKeyDown={preventDefaultEnter}
      onChange={(e) => setClusterSize(parseInt(e.target.value))}
    />
  );
}
