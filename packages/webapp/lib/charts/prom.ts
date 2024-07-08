import { TimeSeries } from "@/lib/charts/types";

// https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries
export interface PrometheusRangeQueryRequestParams {
  query: string;
  start: string | number; // RFC3339 (e.g. 2015-07-01T20:10:30.781Z) or Unix timestamp
  end: string | number; // RFC3339 or Unix timestamp
  step: string | number; // Duration format or float number of seconds
  timeout?: string; // Optional duration
}

export interface PrometheusRangeQueryResponseBase<T> {
  status: string;
  errorType: string;
  error: string;
  data: {
    resultType: string;
    result: Array<{
      metric: Record<string, any>;
      values: T;
    }>;
  };
}

export type PromValues = Array<[number, string]>;
export type PromDateValues = Array<[Date, string]>;

type PrometheusRangeQueryResponseRaw =
  PrometheusRangeQueryResponseBase<PromValues>;
type PrometheusRangeQueryResponse =
  PrometheusRangeQueryResponseBase<PromDateValues>;

export type PromRangeResult = {
  metric: Record<string, any>;
  values: PromDateValues;
}[];

export function promResultToTimeSeries(
  result: PromRangeResult,
  ids: string[]
): TimeSeries[] {
  if (result.length !== ids.length) {
    throw new Error("The length of ids must be equal to the length of result");
  }
  return result.map(({ metric, values }, index) => {
    const id = ids[index]!;
    return {
      id,
      data: values.map(([date, value]) => ({
        x: date,
        y: parseFloat(value),
      })),
    };
  });
}

// define a class that takes in a cluster and when it constructs itself it creates a tmpfile with the kubeconfig for that cluster
export class PrometheusClient {
  private endpoint: string;
  constructor(endpoint: string) {
    this.endpoint = endpoint;
  }

  // rangeQuery: https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries
  async rangeQuery(
    params: PrometheusRangeQueryRequestParams
  ): Promise<PrometheusRangeQueryResponse> {
    const queryString = new URLSearchParams(params as any).toString();
    const response = await fetch(
      `${this.endpoint}/api/v1/query_range?${queryString}`
    );
    const raw = (await response.json()) as PrometheusRangeQueryResponseRaw;
    // convert unix timestamps to dates
    const ret: PrometheusRangeQueryResponse = {
      status: raw.status,
      errorType: raw.errorType,
      error: raw.error,
      data: {
        resultType: raw.data.resultType,
        result: raw.data.result.map(({ metric, values }) => ({
          metric,
          values: values.map(([timestamp, value]) => [
            new Date(timestamp * 1000),
            value,
          ]),
        })),
      },
    };
    return ret;
  }
}

// stepForRange returns the most logical step value for a given time range
// step can be either 1s, 60s, 5m, 10m, 30m, 1h, 3h, 6h, 12h, 1d
// the logic starts with a 1s step and seeing how many steps fit into the range
// if that number is greater than 1440, then we go to the next step size
// we keep going until we get a step size that puts us under 1440 datapoints
const steps = [
  1,
  60,
  5 * 60,
  10 * 60,
  30 * 60,
  60 * 60,
  3 * 60 * 60,
  6 * 60 * 60,
  12 * 60 * 60,
  24 * 60 * 60,
];
export function stepForRange({
  startDate,
  endDate,
}: {
  startDate: Date;
  endDate: Date;
}): number {
  const start = Math.floor(startDate.getTime() / 1000);
  const end = Math.floor(endDate.getTime() / 1000);
  const range = end - start;
  for (const step of steps) {
    if (range / step < 1440) {
      return step;
    }
  }
  return steps[steps.length - 1]!;
}
