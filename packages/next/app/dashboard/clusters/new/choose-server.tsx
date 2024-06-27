import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ServerInfo } from "./shared";
import { cn } from "@/lib/utils";
import { Check } from "lucide-react";
import { forwardRef } from "react";

interface ChooseServerProps {
  data: ServerInfo[];
  serverType: ServerInfo;
  setServerType: (server: ServerInfo) => void;
}

export const ChooseServer = forwardRef(
  (
    { data, serverType, setServerType }: ChooseServerProps,
    ref: React.ForwardedRef<HTMLDivElement>
  ) => {
    return (
      <div className="rounded-md border" ref={ref}>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Type</TableHead>
              <TableHead>CPUs</TableHead>
              <TableHead>Mem (GB)</TableHead>
              <TableHead className="hidden sm:table-cell">Disk (GB)</TableHead>
              <TableHead>Monthly Price</TableHead>
              <TableHead></TableHead>
            </TableRow>
            {data.map((server) => (
              <TableRow
                key={server.name}
                className={cn(
                  "hover:cursor-pointer",
                  serverType.name === server.name && "bg-accent/20"
                )}
                onClick={() => setServerType(server)}
              >
                <TableCell>{server.name}</TableCell>
                <TableCell>{server.cores}</TableCell>
                <TableCell>{server.memory}</TableCell>
                <TableCell className="hidden sm:table-cell">
                  {server.disk}
                </TableCell>
                <TableCell>{server.prettyPriceMonthly}</TableCell>
                <TableCell>
                  {serverType.name === server.name ? (
                    <Check className="w-4 h-4" />
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableHeader>
          <TableBody></TableBody>
        </Table>
      </div>
    );
  }
);

ChooseServer.displayName = "ChooseServer";
