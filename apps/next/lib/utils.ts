import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { KeyboardEvent } from "react";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// preventDefaultEnter prevents enter from submitting the form
export function preventDefaultEnter(event: KeyboardEvent) {
  if (event.key === "Enter") {
    event.preventDefault();
  }
}
