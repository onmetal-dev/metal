"use client";

import { Line, LineChart, CartesianGrid, XAxis, YAxis } from "recharts";

import { ChartContainer } from "@/components/ui/chart";
import { ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { TimeSeries, TimeSeriesDatum } from "@/lib/charts/types";
import { axisTimeFormat } from "@/lib/charts/time";
import dayjs from "dayjs";
import advancedFormat from "dayjs/plugin/advancedFormat";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";

dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.extend(advancedFormat);

interface LineChartProps {
  data: TimeSeries[];
  timeframeSeconds: number;
  yAxis: {
    min: number;
    max: number;
    tickFormatter: (value: any) => string;
    label: string;
  };
  tooltip: {
    xFormatter: (value: Date) => string;
    yFormatter: (value: any) => string;
  };
}

type TransformedTimeSeries = {
  date: Date;
  [seriesId: string]: Date | number;
};

// transform time series into a single array of objects of the form
// { date: Date, [series.id]: number }
// then use the dataKey="date" and dataKey={series.id} to plot the data
// for each series
function transformTimeSeries(data: TimeSeries[]): TransformedTimeSeries[] {
  // check some invariants:
  // must have at least one timeseries
  if (data.length === 0) {
    throw new Error("At least one TimeSeries must be provided");
  }
  // - length of each TimeSeries is the same
  const lengths = data.map((series) => series.data.length);
  if (lengths.some((length) => length !== lengths[0])) {
    throw new Error("All TimeSeries must have the same length");
  }
  // - all TimeSeries have the same x values
  const xValues = data.map((series) => series.data.map((d) => d.x));
  if (
    !xValues.every((series) =>
      series.every((x, i) => x.getTime() === xValues[0]![i]!.getTime())
    )
  ) {
    throw new Error("All TimeSeries must have the same x values");
  }

  let transformedData: TransformedTimeSeries[] = [];
  for (let i = 0; i < data[0]!.data.length; i++) {
    const x = data[0]!.data[i]!.x;
    transformedData.push({
      date: x,
      ...Object.assign(
        {},
        ...data.map((series) => ({
          [series.id]: series.data[i]!.y,
        }))
      ),
    });
  }
  return transformedData;
}

function MetalLineChart({
  data,
  timeframeSeconds,
  yAxis,
  tooltip,
}: LineChartProps) {
  const transformedData = transformTimeSeries(data);
  const firstDataId = data[0]!.id;

  return (
    <ChartContainer config={{}} className="min-h-[200px] w-full h-full">
      <LineChart accessibilityLayer data={transformedData}>
        <XAxis
          dataKey="date"
          tickLine={false}
          tickMargin={15}
          axisLine={false}
          tickFormatter={(date) => axisTimeFormat(date, timeframeSeconds)}
        />
        <YAxis
          dataKey={`${firstDataId}`}
          interval="preserveStartEnd"
          label={{ value: yAxis.label, angle: -90, position: "insideLeft" }}
          tickLine={false}
          tickCount={5}
          tickMargin={15}
          axisLine={false}
          domain={[yAxis.min, yAxis.max]}
          tickFormatter={(y: any) => yAxis.tickFormatter(y)}
        />
        <ChartTooltip
          content={
            <ChartTooltipContent
              indicator="line"
              labelFormatter={(label: any, payload: any[]) => {
                // pull date from first time series
                const d: Date | undefined = payload[0].payload?.date;
                if (!d) {
                  return "";
                }
                return tooltip.xFormatter(d);
              }}
              valueFormatter={(value) => tooltip.yFormatter(value)}
            />
          }
        />
        <CartesianGrid vertical={false} />
        <>
          {data.map((series, idx) => (
            <Line
              key={series.id}
              dataKey={series.id}
              type="monotone"
              stroke={`hsl(var(--chart-${idx + 1}))`}
              strokeWidth={2}
              dot={false}
            />
          ))}
        </>
      </LineChart>
    </ChartContainer>
  );
}

export { MetalLineChart as LineChart };
