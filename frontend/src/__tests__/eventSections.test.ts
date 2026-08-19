import { test, expect } from "bun:test";
import { headlinerName, groupEvents, monthDividers } from "@/utils/eventSections";
import { event } from "./factory";

test("the headliner wins over billing order, and the title is the last resort", () => {
  const e = event({
    name: "Fallback Title",
    artists: [
      { artist_id: "a1", artist_name: "Support", is_headliner: false },
      { artist_id: "a2", artist_name: "Main", is_headliner: true },
    ],
  });
  expect(headlinerName(e)).toBe("Main");

  const noFlag = event({
    artists: [{ artist_id: "a1", artist_name: "First", is_headliner: false }],
  });
  expect(headlinerName(noFlag)).toBe("First");
  expect(headlinerName(event({ name: "Fallback Title" }))).toBe("Fallback Title");
});

test("grouping by artist folds case variants together", () => {
  const groups = groupEvents(
    [
      event({
        id: "1",
        artists: [{ artist_id: "a", artist_name: "Portishead", is_headliner: true }],
      }),
      event({
        id: "2",
        artists: [{ artist_id: "a", artist_name: "portishead", is_headliner: true }],
      }),
    ],
    "artist",
  );

  expect(groups).toHaveLength(1);
  expect(groups[0].events).toHaveLength(2);
});

test("groups are ordered by size", () => {
  const de = (id: string) =>
    event({ id, venue: { name: "V", city: "Berlin", country: "Germany", country_code: "DE" } });
  const groups = groupEvents(
    [
      event({
        id: "x",
        venue: { name: "V", city: "Paris", country: "France", country_code: "FR" },
      }),
      de("1"),
      de("2"),
    ],
    "country",
  );

  expect(groups[0].code).toBe("de");
  expect(groups[0].events).toHaveLength(2);
});

test("dividers appear once per month and mark the year change", () => {
  const rows = monthDividers([
    event({ id: "1", start_date: "2026-11-02" }),
    event({ id: "2", start_date: "2026-11-20" }),
    event({ id: "3", start_date: "2027-01-05" }),
  ]);

  const dividers = rows.filter((r) => r.kind === "divider");
  expect(dividers).toHaveLength(2);
  expect(dividers[0].yearBreak).toBe(false);
  expect(dividers[1].yearBreak).toBe(true);

  // The index on event rows drives row striping, so it counts events, not rows.
  const events = rows.filter((r) => r.kind === "event");
  expect(events.map((r) => (r.kind === "event" ? r.index : -1))).toEqual([0, 1, 2]);
});

test("dividers can be switched off without losing events", () => {
  const rows = monthDividers([event({ id: "1" }), event({ id: "2" })], false);
  expect(rows.every((r) => r.kind === "event")).toBe(true);
  expect(rows).toHaveLength(2);
});
