import { test, expect } from "@playwright/test";
import {
  authed,
  createArtist,
  fetchEvents,
  libraryStatuses,
  radarRows,
  rowId,
  signIn,
  silence,
  token,
} from "./support";

// The whole chain in one sequence: fetch from the stub, see the events on the radar,
// claim one in the browser, refetch a response that no longer carries it, and find it
// still there - delisted, not deleted.
test("a claimed event survives a refetch that drops it", async ({ page, request }) => {
  const t = await token(request);
  const artist = await createArtist(request, t, "Fixture Artist");
  await fetchEvents(request, t, artist.id);

  await signIn(page, request);
  await page.goto("/events");

  const rows = await radarRows(page, "Fixture Artist");
  await expect(rows).toHaveCount(2);

  const kept = await rowId(rows.first());
  await rows.first().getByTestId("claim-going").click();

  await expect
    .poll(async () => (await libraryStatuses(request, t)).find((e) => e.event_id === kept)?.status)
    .toBe("going");

  await silence(request, t, artist.id);
  await fetchEvents(request, t, artist.id);

  expect((await authed(request, t).get(`/api/events/${kept}`)).status()).toBe(200);

  await page.goto(`/events/${kept}`);
  await expect(page.getByTestId("event-lifecycle")).toContainText(/no longer listed/i);
});
