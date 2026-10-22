import { describe, expect, it } from "vitest";

import {
  dateFormatOf,
  formatDate,
  formatDateTime,
  parseTypedDate,
  typingPattern,
} from "@/Utils/dates";

describe("formatDate", () => {
  it("follows the preference", () => {
    expect(formatDate("2026-10-21", "dmy")).toBe("21/10/2026");
    expect(formatDate("2026-10-21", "mdy")).toBe("10/21/2026");
    expect(formatDate("2026-10-21", "ymd")).toBe("2026-10-21");
    expect(formatDate("2026-10-21", "long")).toMatch(/^21 \w+ 2026$/);
  });

  it("defaults to day/month/year and survives bad input", () => {
    expect(formatDate("2026-10-21")).toBe("21/10/2026");
    expect(formatDate("2026-10-21", "dd.mm")).toBe("21/10/2026");
    expect(formatDate(null)).toBe("");
    expect(formatDate("not a date")).toBe("");
  });

  it("adds the time", () => {
    expect(formatDateTime("2026-10-21T15:56:00", "ymd")).toBe(
      "2026-10-21 15:56",
    );
    expect(formatDateTime(undefined)).toBe("");
  });
});

describe("parseTypedDate", () => {
  it("reads the person's own format", () => {
    expect(parseTypedDate("21/10/2026", "dmy")).toBe("2026-10-21");
    expect(parseTypedDate("10/21/2026", "mdy")).toBe("2026-10-21");
    expect(parseTypedDate("2026-10-21", "ymd")).toBe("2026-10-21");
    expect(parseTypedDate("21/10/2026", "long")).toBe("2026-10-21");
  });

  it("refuses impossible or ambiguous input", () => {
    expect(parseTypedDate("31/02/2026", "dmy")).toBeNull();
    expect(parseTypedDate("10/21/2026", "dmy")).toBeNull();
    expect(parseTypedDate("21-10-2026", "dmy")).toBeNull();
    expect(parseTypedDate("", "dmy")).toBeNull();
  });
});

it("typingPattern and dateFormatOf", () => {
  expect(typingPattern("mdy")).toBe("MM/DD/YYYY");
  expect(typingPattern("long")).toBe("DD/MM/YYYY");
  expect(dateFormatOf(undefined)).toBe("dmy");
});
