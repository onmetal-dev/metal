import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function NoClusters() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-2 xl:grid-cols-4">
      <Card className="sm:col-span-2">
        <CardHeader className="pb-3">
          <CardTitle>Clusters</CardTitle>
          <CardDescription className="max-w-lg text-balance leading-relaxed">
            Set up a cluster to begin deploying applications.
          </CardDescription>
        </CardHeader>
        <CardFooter>
          <Button asChild>
            <Link href="/dashboard/clusters/new">Create New Cluster</Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
