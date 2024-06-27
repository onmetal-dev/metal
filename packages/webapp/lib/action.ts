export type ServerActionState = {
  message: string;
  isError?: boolean;
};

export const serverActionInitialState: ServerActionState = {
  message: "",
  isError: false,
};
