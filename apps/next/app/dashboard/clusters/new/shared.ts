export type createHetznerClusterState = {
  message: string;
  isError?: boolean;
};

export const createHetznerClusterInitialState: createHetznerClusterState = {
  message: "",
  isError: false,
};

export type ServerInfo = {
  name: string;
  cores: number;
  memory: number;
  disk: number;
  priceHourly: string;
  priceMonthly: string;
  prettyPriceHourly: string;
  prettyPriceMonthly: string;
  currency: string;
};
