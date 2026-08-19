import { test, expect } from "@playwright/test";
import {
  authed,
  createArtist,
  fetchEvents,
  radarRows,
  rowId,
  signIn,
  silence,
  token,
} from "./support";

// The radar is forward-looking discovery, not a list of every row in the database: a
// claim narrows it, and a listing that has gone away is hidden until asked for.
test("the radar filters follow the claim and the listing state", async ({ page, request }) => {
  const t = await token(request);
  const artist = await createArtist(request, t, "Filter Artist");
  await fetchEvents(request, t, artist.id);

  await signIn(page, request);
  await page.goto("/events");

  const rows = await radarRows(page, "Filter Artist");
  await expect(rows).toHaveCount(2);

  const claimed = await rowId(rows.first());
  await authed(request, t).put(`/api/events/${claimed}/status`, { status: "interested" });
  await page.reload();
  await radarRows(page, "Filter Artist");

  await page.getByTestId("filter-only-claimed").click();
  await expect(rows).toHaveCount(1);
  await page.getByTestId("filter-only-claimed").click();
  await expect(rows).toHaveCount(2);

  // The feed drops both: the un-claimed one is swept, the claimed one is delisted.
  await silence(request, t, artist.id);
  await fetchEvents(request, t, artist.id);
  await page.reload();
  await radarRows(page, "Filter Artist");

  await expect(rows).toHaveCount(0);
  await page.getByTestId("filter-show-past").click();
  await expect(rows).toHaveCount(1);
});
