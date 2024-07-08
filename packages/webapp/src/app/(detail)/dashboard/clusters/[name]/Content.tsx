"use client";
import { LineChart } from "@/components/charts/LineChart";
import { useCallback, useEffect, useRef, useState } from "react";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { fetchClusterMetrics } from "./actions";
import { TimeSeries } from "@/lib/charts/types";
import { useRouter } from "next/navigation";
import { useKeyPressEvent } from "react-use";
import prettyBytes from "pretty-bytes";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { KeySymbol, ShortcutsProvider } from "@/components/ui/keyboard";
import Help from "./Help";
import { useCommandItems } from "@/components/CommandMenu";
import { minMaxForTimeSeries, tooltipTimeFormat } from "@/lib/charts/time";

const timeframes = [
  {
    label: "1H",
    seconds: 60 * 60,
  },
  {
    label: "1D",
    seconds: 60 * 60 * 24,
  },
  {
    label: "1W",
    seconds: 60 * 60 * 24 * 7,
  },
];

export default function Content({
  clusterName,
  initialData,
}: {
  clusterName: string;
  initialData: {
    cpu: TimeSeries[];
    mem: TimeSeries[];
    cpuRequests: TimeSeries[];
    memRequests: TimeSeries[];
  };
}) {
  const [timeframe, setTimeframe] = useState(timeframes[0]!);
  const [data, setData] = useState(initialData);
  const [dataTimeframe, setDataTimeframe] = useState(timeframes[0]!);
  const metricCharts = [
    {
      title: "CPU Utilization",
      data: data.cpu,
      yAxis: {
        ...minMaxForTimeSeries({
          ts: data.cpu,
          xOrY: "y",
          maxMin: 0,
          minMax: 100,
        }),
        label: "%",
        tickFormatter: (value: any) => `${value}`,
      },
      tooltip: {
        xFormatter: (value: Date) => tooltipTimeFormat(value),
        yFormatter: (value: any) => `${value.toFixed(1)}%`,
      },
    },
    {
      title: "Memory Utilization",
      data: data.mem,
      yAxis: {
        ...minMaxForTimeSeries({
          ts: data.mem,
          xOrY: "y",
          maxMin: 0,
          minMax: 100,
        }),
        tickFormatter: (value: any) => `${value}`,
        label: "%",
      },
      tooltip: {
        xFormatter: (value: Date) => tooltipTimeFormat(value),
        yFormatter: (value: any) => `${value.toFixed(1)}%`,
      },
    },
    {
      title: "CPU Requests",
      data: data.cpuRequests,
      yAxis: {
        ...minMaxForTimeSeries({
          ts: data.cpuRequests,
          xOrY: "y",
          maxMin: 0,
          minMax: 10,
        }),
        tickFormatter: (value: any) => `${value}`,
        label: "# cpu",
      },
      tooltip: {
        xFormatter: (value: Date) => tooltipTimeFormat(value),
        yFormatter: (value: any) => `${value.toFixed(2)}`,
      },
    },
    {
      title: "Memory Requests",
      data: data.memRequests,
      yAxis: {
        ...minMaxForTimeSeries({
          ts: data.memRequests,
          xOrY: "y",
          maxMin: 0,
          minMax: 10,
        }),
        tickFormatter: (value: any) => prettyBytes(value),
        label: "",
      },
      tooltip: {
        xFormatter: (value: Date) => tooltipTimeFormat(value),
        yFormatter: (value: any) => prettyBytes(value),
      },
    },
    // todo: network, disk
  ];

  const handleTimeframeChange = useCallback(
    (timeframe: (typeof timeframes)[number]) => {
      fetchClusterMetrics({
        timeframeSeconds: timeframe.seconds,
        clusterName,
      }).then((data) => {
        setData(data);
        setDataTimeframe(timeframe);
      });
      setTimeframe(timeframe);
    },
    [clusterName]
  );
  useKeyPressEvent("[", () => {
    const idx = timeframes.findIndex((t) => t.label === timeframe.label);
    const newTimeframe = timeframes[Math.max(0, idx - 1)]!;
    handleTimeframeChange(newTimeframe);
  });
  useKeyPressEvent("]", () => {
    const idx = timeframes.findIndex((t) => t.label === timeframe.label);
    const newTimeframe = timeframes[Math.min(timeframes.length - 1, idx + 1)]!;
    handleTimeframeChange(newTimeframe);
  });
  const [helpOpen, setHelpOpen] = useState(false);

  useKeyPressEvent("?", () => {
    setHelpOpen(true);
  });

  const router = useRouter();
  // this is incredibly hacky, but this solves for helpOpen state being false despite just closing the help dialog (state ordering?)
  const helpMenuLastClosedTime = useRef<null | Date>(null);
  useEffect(() => {
    if (!helpOpen) {
      helpMenuLastClosedTime.current = new Date();
    }
  }, [helpOpen]);
  useKeyPressEvent("Escape", () => {
    if (
      !helpOpen &&
      helpMenuLastClosedTime.current &&
      Date.now() - helpMenuLastClosedTime.current.getTime() > 500
    ) {
      router.push("/dashboard/clusters");
    } else {
      setHelpOpen(false);
    }
  });
  const { addCommandItem, removeCommandItem, setGroupPriority } =
    useCommandItems();
  useEffect(() => {
    setGroupPriority("Cluster Detail Page Actions", 99);
    addCommandItem({
      group: "Cluster Detail Page Actions",
      label: "Back to Clusters",
      onSelect: () => {
        router.push("/dashboard/clusters");
      },
    });
    return () => {
      removeCommandItem("Back to Clusters");
    };
  }, [addCommandItem, removeCommandItem, setGroupPriority, router]);

  return (
    <TooltipProvider>
      <ShortcutsProvider>
        <Help open={helpOpen} setOpen={setHelpOpen} />
        <div className="w-full bg-muted/40 max-w-7xl h-full top-20 rounded-t-md shadow-2xl left-0 right-0 mx-auto bg-scroll">
          <div
            id="Blocks"
            className="px-20 mx-auto w-full h-full flex flex-column"
            style={{ minHeight: "calc(100vh - 32px)", willChange: "transform" }}
          >
            <div id="Sections" className="w-full max-w-4xl pt-4 mx-auto">
              <section id="metrics" className="mb-14">
                <div className="flex items-center w-full justify-end mb-4">
                  <Tooltip>
                    <TooltipContent side="bottom">
                      <div>
                        <span className="mr-2 text-xs">Change timeframe</span>
                        <KeySymbol disableTooltip={true} keyName="[" />
                        <KeySymbol disableTooltip={true} keyName="]" />
                      </div>
                    </TooltipContent>
                    <TooltipTrigger asChild>
                      <ToggleGroup
                        type="single"
                        value={timeframe.label}
                        className="text-muted-foreground"
                        disabled={timeframe.label !== dataTimeframe.label}
                      >
                        {timeframes.map((timeframe) => (
                          <ToggleGroupItem
                            key={timeframe.label}
                            value={timeframe.label}
                            aria-label={timeframe.label}
                            className="cursor-pointer"
                            onClick={() => {
                              handleTimeframeChange(timeframe);
                            }}
                          >
                            <span>{timeframe.label}</span>
                          </ToggleGroupItem>
                        ))}
                      </ToggleGroup>
                    </TooltipTrigger>
                  </Tooltip>
                </div>
                <div className="grid gap-6 grid-cols-1 lg:grid-cols-2">
                  {metricCharts.map(({ title, data, yAxis, tooltip }) => (
                    <div
                      key={title}
                      className={`h-[300px] bg-background rounded-sm horizontal center flex flex-col p-4 text-muted-foreground ${
                        timeframe.label !== dataTimeframe.label ? "blur-sm" : ""
                      }`}
                    >
                      <h4 className="text-sm font-medium pb-4">{title}</h4>
                      <LineChart
                        data={data}
                        timeframeSeconds={dataTimeframe.seconds}
                        yAxis={yAxis}
                        tooltip={tooltip}
                      />
                    </div>
                  ))}
                </div>
              </section>
            </div>
          </div>
        </div>
      </ShortcutsProvider>
    </TooltipProvider>
  );
}
