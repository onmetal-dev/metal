"use client";
import { ResponsiveLine, SliceTooltipProps, Serie } from "@nivo/line";
import React from "react";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import advancedFormat from "dayjs/plugin/advancedFormat";

dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.extend(advancedFormat);

interface LineChartProps {
  data: readonly Serie[];
  timeframeSeconds: number;
  yAxis: {
    min: number;
    max: number;
    format: (value: any) => string;
    legend: string;
  };
  tooltip: {
    yFormat: (value: any) => string;
  };
}

const CustomXAxisSliceTooltip: React.FunctionComponent<SliceTooltipProps> = ({
  slice,
}) => {
  return (
    <div className="z-50 w-fit flex flex-col overflow-hidden rounded-md border bg-popover px-3 py-1.5 text-xs text-muted-foreground shadow-md backdrop-blur">
      <div>{slice.points[0]!.data.xFormatted}</div>
      {slice.points.map((point) => (
        <div key={point.id} className="flex items-center gap-2">
          <span>
            <svg
              width="3"
              height="13"
              viewBox="0 0 3 13"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <rect
                width="3"
                height="13"
                rx="1.5"
                fill={`url(#linear_${point.color})`}
              ></rect>
              <defs>
                <linearGradient
                  id={`linear_${point.color}`}
                  x1="0%"
                  y1="0%"
                  x2="100%"
                  y2="100%"
                  gradientUnits="userSpaceOnUse"
                >
                  <stop
                    offset="0%"
                    style={{
                      stopColor: point.color,
                      stopOpacity: 0.8,
                    }}
                  />
                  <stop
                    offset="50%"
                    style={{
                      stopColor: point.color,
                      stopOpacity: 1.0,
                    }}
                  />
                  <stop
                    offset="100%"
                    style={{
                      stopColor: point.color,
                      stopOpacity: 0.8,
                    }}
                  />
                </linearGradient>
              </defs>
            </svg>
          </span>
          <strong>{point.data.yFormatted}</strong>
          <span className="text-muted-foreground">{point.serieId}</span>
        </div>
      ))}
    </div>
  );
};

function timeFormat(timeframeSeconds: number) {
  if (timeframeSeconds <= 86400 /* 1D */) {
    return "HH:mm";
  } else if (timeframeSeconds <= 604800 /* 1W */) {
    return "MMM D, HH:mm";
  }
  return "MMM D, HH:mm";
}

function timeTickValues(timeframeSeconds: number) {
  if (timeframeSeconds <= 3600) {
    return "every 10 minutes";
  } else if (timeframeSeconds <= 86400) {
    return "every 3 hours";
  } else if (timeframeSeconds <= 604800) {
    return "every 1 day";
  }
  return "every 1 day";
}

const LineChart = ({
  data,
  timeframeSeconds,
  yAxis,
  tooltip,
}: LineChartProps) => {
  return (
    <ResponsiveLine
      animate={true}
      axisBottom={{
        tickValues: timeTickValues(timeframeSeconds),
        tickSize: 5,
        tickPadding: 5,
        tickRotation: 0,
        truncateTickAt: 0,
        format: (value: any) => {
          return dayjs(value).format(timeFormat(timeframeSeconds));
        },
      }}
      axisLeft={{
        tickValues: 5,
        tickSize: 5,
        tickPadding: 5,
        tickRotation: 0,
        legend: "%",
        legendOffset: -35,
        legendPosition: "middle",
        truncateTickAt: 0,
        format: yAxis.format,
      }}
      axisRight={null}
      axisTop={null}
      data={data}
      enableCrosshair={false}
      enableGridX={false}
      enableGridY={true}
      enablePoints={false}
      enableSlices={"x"}
      enableTouchCrosshair={false}
      legends={[]}
      lineWidth={2}
      margin={{
        top: 20,
        right: 50,
        bottom: 50,
        left: 50,
      }}
      useMesh={true}
      yFormat={tooltip.yFormat}
      theme={{
        background: "transparent",
        axis: {
          domain: {
            line: {
              stroke: "transparent",
              strokeWidth: 1,
            },
          },
          ticks: {
            line: {
              stroke: "transparent",
              strokeWidth: 0,
            },
            text: {
              fill: "#868F97",
              fontSize: 10,
              fontWeight: 500,
              fillOpacity: 0.8,
            },
          },
        },
        grid: {
          line: {
            stroke: "#868F97",
            strokeWidth: 1,
            strokeDasharray: "3 3",
            strokeOpacity: 0.3,
          },
        },
      }}
      sliceTooltip={CustomXAxisSliceTooltip}
      xFormat={(value: any) => {
        return dayjs(value).format("MMM D, HH:mm z");
      }}
      xScale={{ type: "time" }}
      yScale={{
        type: "linear",
        min: yAxis.min,
        max: yAxis.max,
        stacked: false,
        reverse: false,
      }}
    />
  );
};

export default LineChart;
