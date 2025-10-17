import { describe, expect, it } from "vitest";

import {
  formatDateTime,
  formatDateTimeString,
  formatDuration,
  humanFileSize,
} from "../format";

describe("formatDateTime", () => {
  it("formats a date object with time", () => {
    expect(formatDateTime(new Date("2023-01-15T12:30:45"))).toBe(
      "2023-01-15 12:30:45",
    );
  });

  it("formats a date string with time", () => {
    expect(formatDateTime(new Date("2023-01-15T12:30:45"))).toBe(
      "2023-01-15 12:30:45",
    );
  });

  it("returns an empty string for undefined input", () => {
    expect(formatDateTime(undefined)).toBe("");
  });
});

describe("formatDateTimeString", () => {
  it("formats a valid date string", () => {
    expect(formatDateTimeString("2023-01-15T12:30:45")).toBe(
      "2023-01-15 12:30:45",
    );
  });

  it("returns 'Invalid date' for an invalid date string", () => {
    expect(formatDateTimeString("invalid-date")).toBe("Invalid date");
  });
});

describe("formatDuration", () => {
  it("formats duration between two dates", () => {
    const from = new Date("2023-01-15T12:00:00Z");
    const to = new Date("2023-01-15T14:03:00Z");
    expect(formatDuration(from, to)).toBe("2 hours");
  });

  it("handles durations less than a minute", () => {
    const from = new Date("2023-01-15T12:00:00Z");
    const to = new Date("2023-01-15T12:00:30Z");
    expect(formatDuration(from, to)).toBe("a few seconds");
  });
});

describe("humanFileSize", () => {
  it("formats bytes", () => {
    expect(humanFileSize(123)).toBe("123 bytes");
  });

  it("formats kilobytes to a whole number", () => {
    expect(humanFileSize(1500)).toBe("2 KB");
  });

  it("formats kilobytes to one decimal", () => {
    expect(humanFileSize(1500, 1)).toBe("1.5 KB");
  });

  it("formats megabytes to two decimals", () => {
    expect(humanFileSize(1520000, 2)).toBe("1.52 MB");
  });

  it("formats gigabytes to one decimal", () => {
    expect(humanFileSize(1500000000, 1)).toBe("1.5 GB");
  });

  it("formats terabytes to one decimal", () => {
    expect(humanFileSize(1500000000000, 1)).toBe("1.5 TB");
  });

  it("returns '0 bytes' for zero", () => {
    expect(humanFileSize(0)).toBe("0 bytes");
  });
});
