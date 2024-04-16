import Link from "next/link";
import {
  Blocks,
  ChevronLeft,
  ChevronRight,
  Copy,
  CreditCard,
  Database,
  File,
  Link2,
  ListFilter,
  MoreVertical,
  Package2,
  PanelLeft,
  PanelsTopLeft,
  Search,
  Server,
  Settings,
  Settings2,
  SettingsIcon,
  Truck,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
} from "@/components/ui/pagination";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipProvider,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Logo } from "@/components/Logo";
import { UserButton } from "@clerk/nextjs";
import { SidebarNav } from "@/components/SidebarNav";
import { SidebarNavSm } from "@/components/SidebarNavSm";
import { Topbar } from "@/components/Topbar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40">
      <SidebarNav />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14">
        <Topbar />
        <main className="flex-1 items-start p-4 sm:px-6 sm:py-0">
          {children}
        </main>
      </div>
    </div>
  );
}
