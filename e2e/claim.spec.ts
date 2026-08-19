import { test, expect } from "@playwright/test";
import { authed, createEvent, daysFromNow, signIn, token, uniq } from "./support";

test("a claim made through the API shows up in the diary", async ({ page, request }) => {
  const t = await token(request);
  const event = await createEvent(request, t, {
    name: `Claimed ${uniq()}`,
    start_date: daysFromNow(-20),
    venue: { name: "Alte Halle", city: "Koeln", country_code: "de" },
    status: "attended",
  });

  await signIn(page, request);
  await page.goto("/library");
  await page.getByTestId("library-tab-attended").click();

  await expect(page.getByTestId(`library-row-${event.id}`)).toBeVisible();
});

test("a claimed event refuses to be deleted and stays on its page", async ({ page, request }) => {
  const t = await token(request);
  const event = await createEvent(request, t, {
    name: `Undeletable ${uniq()}`,
    start_date: daysFromNow(31),
    status: "going",
  });

  const res = await authed(request, t).delete(`/api/events/${event.id}`);
  expect(res.status(), "a claim outranks a delete").toBe(409);

  await signIn(page, request);
  await page.goto(`/events/${event.id}`);
  await expect(page.getByTestId("event-detail-name")).toHaveText(event.name);
});
