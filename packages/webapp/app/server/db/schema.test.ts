import { describe, expect, test } from "bun:test";
import { ZodError } from "zod";
import { mustParseCpu, mustParseMemory, resourcesSchema } from "./schema";

describe("resourcesSchema", () => {
  const validCpuCases = [
    { input: "100m", expected: 0.1 },
    { input: "0.1", expected: 0.1 },
    { input: ".25", expected: 0.25 },
    { input: "1", expected: 1 },
    { input: "1.234", expected: 1.234 },
  ];

  const invalidCpuCases = [
    { input: "100.1m", error: "Millicpu requests must be integers." },
    { input: "0.0001", error: "Maximum precision is 0.001." },
    { input: "abc", error: "abc is not a valid CPU request." },
  ];

  const validMemoryCases = [
    { input: "1000", expected: 1000 },
    { input: "1M", expected: 1000 ** 2 },
    { input: "1Mi", expected: 1024 ** 2 },
    { input: "1.5G", expected: 1.5 * 1000 ** 3 },
  ];

  const invalidMemoryCases = [
    {
      input: "1Z",
      error: "1Z is not a valid memory request",
    },
    {
      input: "abc",
      error: "abc is not a valid memory request",
    },
    {
      input: "1.1",
      error:
        "If not specifying a unit, memory requests are interpreted as bytes and must be integers.",
    },
  ];

  validCpuCases.forEach(({ input, expected }) => {
    test(`converts valid CPU input: ${input}`, () => {
      const num = mustParseCpu(input);
      expect(num).toBe(expected);
    });
    test(`valid CPU input: ${input}`, () => {
      expect(() =>
        resourcesSchema.parse({ cpu: input, memory: "256M" })
      ).not.toThrow();
    });
  });

  invalidCpuCases.forEach(({ input, error }) => {
    test(`invalid CPU input: ${input}`, () => {
      expect(() => {
        try {
          resourcesSchema.parse({ cpu: input, memory: "256M" });
        } catch (e: any) {
          const error = e as ZodError;
          throw error.issues[0]!.message;
        }
      }).toThrow(error);
    });
  });

  validMemoryCases.forEach(({ input, expected }) => {
    test(`converts valid memory input: ${input}`, () => {
      const num = mustParseMemory(input);
      expect(num).toBe(expected);
    });
    test(`valid memory input: ${input}`, () => {
      expect(() =>
        resourcesSchema.parse({ cpu: "100m", memory: input })
      ).not.toThrow();
    });
  });

  invalidMemoryCases.forEach(({ input, error }) => {
    test(`invalid memory input: ${input}`, () => {
      expect(() => {
        try {
          resourcesSchema.parse({ cpu: "100m", memory: input });
        } catch (e: any) {
          const error = e as ZodError;
          throw error.issues[0]!.message;
        }
      }).toThrow(error);
    });
  });
});
