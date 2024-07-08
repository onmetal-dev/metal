export type TimeSeries = {
  id: string;
  data: TimeSeriesDatum[];
};

export type TimeSeriesDatum = {
  x: Date;
  y: number;
};
