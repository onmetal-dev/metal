export type createHetznerProjectState = {
  message: string;
  isError?: boolean;
};

export const createHetznerProjectInitialState: createHetznerProjectState = {
  message: "",
  isError: false,
};
