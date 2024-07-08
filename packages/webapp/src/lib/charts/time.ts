import dayjs from "dayjs";
import advancedFormat from "dayjs/plugin/advancedFormat";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import { TimeSeries } from "./types";

dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.extend(advancedFormat);

// axisTimeFormat formats the axis text for a date value
export function axisTimeFormat(
  date: Date,
  timeframeSeconds: number,
  withTz: boolean = false
): string {
  let tzFormat = withTz ? " z" : "";
  let format = `MMM D, HH:mm${tzFormat}`;
  if (timeframeSeconds <= 86400 /* 1D */) {
    format = `HH:mm${tzFormat}`;
  } else if (timeframeSeconds <= 604800 /* 1W */) {
    format = `MMM D${tzFormat}`;
  }
  return dayjs(date).format(format);
}

// tooltipTimeFormat should provide more detail since the user wants info for a specific point
export function tooltipTimeFormat(date: Date): string {
  return dayjs(date).format();
}

// minMaxForTimeSeries returns the min and max values across a bunch of time series
export function minMaxForTimeSeries({
  ts,
  xOrY,
  minMax,
  maxMin,
}: {
  ts: TimeSeries[];
  xOrY: "x" | "y";
  maxMin: number;
  minMax: number;
}) {
  let min = null;
  let max = null;
  for (const series of ts) {
    for (const datum of series.data) {
      const value = datum[xOrY] as number;
      if (value != null) {
        if (min === null || value < min) min = value;
        if (max === null || value > max) max = value;
      }
    }
  }
  min = Math.min(min ?? 0, maxMin);
  max = Math.max(max ?? 0, minMax);
  return { min, max };
}
