export type serverActionState = {
  message: string;
  isError?: boolean;
};

export const serverActionInitialState: serverActionState = {
  message: "",
  isError: false,
};
