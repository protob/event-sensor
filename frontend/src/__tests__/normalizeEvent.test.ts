import { test, expect } from "bun:test";
import { normalizeEvent } from "@/api/events";
import { event } from "./factory";

// artists = performances the catalog knows; lineup = everyone on the bill. A name-only
// act must not reach the artist links, or the UI would render a dead link.
test("performances split into catalog artists and the full bill", () => {
  const e = normalizeEvent(
    event({
      performances: [
        { artist_id: "a1", artist_name: "Known", is_headliner: true },
        { artist_id: null, artist_name: "Name Only", is_headliner: false },
      ],
    }),
  );

  expect(e.artists?.map((a) => a.artist_name)).toEqual(["Known"]);
  expect(e.lineup?.map((l) => l.artist_name)).toEqual(["Known", "Name Only"]);
});

test("an event with no performances gets empty lists, not undefined", () => {
  const e = normalizeEvent(event({ performances: undefined }));
  expect(e.artists).toEqual([]);
  expect(e.lineup).toEqual([]);
});

test("lineup keys are unique inside one event", () => {
  const e = normalizeEvent(
    event({
      id: "evt",
      performances: [
        { artist_id: null, artist_name: "Same", is_headliner: false },
        { artist_id: null, artist_name: "Same", is_headliner: false },
      ],
    }),
  );

  const keys = e.lineup?.map((l) => l.id) ?? [];
  expect(new Set(keys).size).toBe(keys.length);
});
