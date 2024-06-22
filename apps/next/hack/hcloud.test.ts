import { describe, expect, it } from "bun:test";
import { RoundingDirection, hoursBetween, roundToNearestHour } from "./hcloud";

describe("hoursBetween", () => {
  it("calculates hours between two dates on the same day", () => {
    const date1 = new Date("2023-10-01T00:00:00Z");
    const date2 = new Date("2023-10-01T12:00:00Z");
    expect(hoursBetween(date1, date2)).toBe(12);
  });

  it("calculates hours between two dates on different days", () => {
    const date1 = new Date("2023-10-01T00:00:00Z");
    const date2 = new Date("2023-10-02T00:00:00Z");
    expect(hoursBetween(date1, date2)).toBe(24);
  });

  it("calculates hours between two dates with minutes and seconds", () => {
    const date1 = new Date("2023-10-01T00:00:00Z");
    const date2 = new Date("2023-10-01T01:30:00Z");
    expect(hoursBetween(date1, date2)).toBe(1.5);
  });

  it("returns a negative value if the first date is after the second date", () => {
    const date1 = new Date("2023-10-02T00:00:00Z");
    const date2 = new Date("2023-10-01T00:00:00Z");
    expect(hoursBetween(date1, date2)).toBe(-24);
  });

  it("returns 0 if the dates are the same", () => {
    const date1 = new Date("2023-10-01T00:00:00Z");
    const date2 = new Date("2023-10-01T00:00:00Z");
    expect(hoursBetween(date1, date2)).toBe(0);
  });
});

describe("roundToNearestHour", () => {
  it("should round up to the nearest hour", () => {
    const date = new Date("2023-10-01T10:30:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Up);
    expect(roundedDate).toEqual(new Date("2023-10-01T11:00:00Z"));
  });

  it("should round down to the nearest hour", () => {
    const date = new Date("2023-10-01T10:30:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Down);
    expect(roundedDate).toEqual(new Date("2023-10-01T09:00:00Z"));
  });

  it("should not change the date if it is already at the start of the hour and rounding up", () => {
    const date = new Date("2023-10-01T10:00:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Up);
    expect(roundedDate).toEqual(new Date("2023-10-01T11:00:00Z"));
  });

  it("should not change the date if it is already at the start of the hour and rounding down", () => {
    const date = new Date("2023-10-01T10:00:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Down);
    expect(roundedDate).toEqual(new Date("2023-10-01T09:00:00Z"));
  });

  it("should handle rounding up at the end of the day", () => {
    const date = new Date("2023-10-01T23:30:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Up);
    expect(roundedDate).toEqual(new Date("2023-10-02T00:00:00Z"));
  });

  it("should handle rounding down at the start of the day", () => {
    const date = new Date("2023-10-01T00:30:00Z");
    const roundedDate = roundToNearestHour(date, RoundingDirection.Down);
    expect(roundedDate).toEqual(new Date("2023-09-30T23:00:00Z"));
  });
});
