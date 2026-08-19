import { test, expect } from "bun:test";
import {
  isRadarEvent,
  matchesQuery,
  inDateRange,
  sortEvents,
  withinDays,
} from "@/composables/useEventFilters";
import { event } from "./factory";

test("the radar shows only future listed Ticketmaster events", () => {
  expect(isRadarEvent(event(), false)).toBe(true);
  expect(isRadarEvent(event({ source: "manual" }), false)).toBe(false);
  expect(isRadarEvent(event({ is_past: true }), false)).toBe(false);
  expect(isRadarEvent(event({ listing_state: "delisted" }), false)).toBe(false);
});

test("show-past reveals past and delisted events but not manual ones", () => {
  expect(isRadarEvent(event({ is_past: true }), true)).toBe(true);
  expect(isRadarEvent(event({ listing_state: "cancelled" }), true)).toBe(true);
  expect(isRadarEvent(event({ source: "manual" }), true)).toBe(false);
});

test("search reaches the lineup, not only the title", () => {
  const e = event({
    name: "Summer Nights",
    venue: { name: "Arena", city: "Berlin", country: "Germany", country_code: "DE" },
    artists: [{ artist_id: "a1", artist_name: "Kruder", is_headliner: true }],
  });

  expect(matchesQuery(e, "kruder")).toBe(true);
  expect(matchesQuery(e, "BERLIN")).toBe(true);
  expect(matchesQuery(e, "")).toBe(true);
  expect(matchesQuery(e, "nothing here")).toBe(false);
});

test("the date range compares days, so a time part never excludes its own day", () => {
  const e = event({ start_date: "2026-06-01T22:00:00Z" });
  expect(inDateRange(e, "2026-06-01", "2026-06-01")).toBe(true);
  expect(inDateRange(e, "2026-06-02", "")).toBe(false);
  expect(inDateRange(e, "", "")).toBe(true);
});

test("sorting returns a new array and leaves the input alone", () => {
  const input = [
    event({ id: "b", start_date: "2026-06-02" }),
    event({ id: "a", start_date: "2026-06-01" }),
  ];
  const sorted = sortEvents(input, "date-asc");

  expect(sorted.map((e) => e.id)).toEqual(["a", "b"]);
  expect(input.map((e) => e.id)).toEqual(["b", "a"]);
});

test("the day window includes today and excludes what is past it", () => {
  const inDays = (n: number) => new Date(Date.now() + n * 86400_000).toISOString();

  expect(withinDays(event({ start_date: inDays(1) }), 7)).toBe(true);
  expect(withinDays(event({ start_date: inDays(30) }), 7)).toBe(false);
  expect(withinDays(event({ start_date: inDays(-3) }), 7)).toBe(false);
  expect(withinDays(event({ start_date: "not a date" }), 7)).toBe(false);
});
