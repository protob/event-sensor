import { test, expect, describe } from "bun:test";
import { isDateOnly, parseLocal, formatISO, formatEventDate } from "@/utils/dates";

describe("date-only values", () => {
  test("are recognised by the absence of a time part", () => {
    expect(isDateOnly("2026-06-01")).toBe(true);
    expect(isDateOnly("2026-06-01T20:00:00Z")).toBe(false);
  });

  // new Date("2026-06-01") is midnight UTC, which is the previous evening west of
  // Greenwich. Parsing from the parts keeps the day the API sent.
  test("keep their day regardless of the local timezone", () => {
    const d = parseLocal("2026-06-01");
    expect(d.getFullYear()).toBe(2026);
    expect(d.getMonth()).toBe(5);
    expect(d.getDate()).toBe(1);
  });

  test("round-trip through formatISO unchanged", () => {
    expect(formatISO("2026-06-01")).toBe("2026-06-01");
    expect(formatISO("2026-06-01", true)).toBe("26-06-01");
  });

  test("are rendered without a time", () => {
    expect(formatEventDate("2026-06-01")).not.toMatch(/\d\d:\d\d/);
    expect(formatEventDate("2026-06-01T20:30:00Z")).toMatch(/\d\d:\d\d/);
  });
});

test("unparseable input never leaks NaN into the UI", () => {
  expect(formatEventDate("not a date")).toBe("Invalid Date");
  expect(formatISO("not a date")).not.toContain("NaN");
});
