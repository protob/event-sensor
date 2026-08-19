import { test, expect } from "@playwright/test";
import {
  createArtist,
  fetchEvents,
  libraryStatuses,
  radarRows,
  rowId,
  signIn,
  token,
} from "./support";

// Claiming a festival weekend one row at a time is the thing the bulk bar exists to
// avoid: a shift-range plus one menu choice has to claim every row in the range.
test("a shift-selected range is claimed in one action", async ({ page, request }) => {
  const t = await token(request);
  const artist = await createArtist(request, t, "Radar Artist");
  await fetchEvents(request, t, artist.id);

  await signIn(page, request);
  await page.goto("/events");

  const rows = await radarRows(page, "Radar Artist");
  await expect(rows).toHaveCount(4);

  await rows.nth(0).getByTestId("event-select").click();
  await rows.nth(2).getByTestId("event-select").click({ modifiers: ["Shift"] });
  await expect(page.getByTestId("bulk-count")).toContainText("3");

  const picked = await Promise.all([0, 1, 2].map((i) => rowId(rows.nth(i))));

  await page.getByTestId("bulk-set-status").click();
  await page.getByTestId("bulk-status-interested").click();

  await expect
    .poll(async () => {
      const entries = await libraryStatuses(request, t);
      return entries.filter((e) => picked.includes(e.event_id) && e.status === "interested").length;
    })
    .toBe(3);
});
