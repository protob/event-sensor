import { test, expect } from "@playwright/test";
import { daysFromNow, eventIdByName, optionSelect, signIn, token, uniq } from "./support";

// A show nobody sold tickets for online still belongs in the diary, so the manual form
// has to reach the same library the fetched events land in.
test("a past show entered by hand lands in the diary", async ({ page, request }) => {
  const t = await token(request);
  const name = `Basement Show ${uniq()}`;

  await signIn(page, request);
  await page.goto("/library");
  await page.getByTestId("library-add-event").click();

  await page.getByTestId("manual-name").fill(name);
  await page.getByTestId("manual-start-date").fill(daysFromNow(-40));
  await page.getByTestId("manual-venue-name").fill("Klub Pod Schodami");
  await page.getByTestId("manual-venue-city").fill("Krakow");
  await optionSelect(page, "manual-status-attended").selectOption("attended");
  await page.getByTestId("manual-save").click();

  const id = await eventIdByName(request, t, name);
  await page.getByTestId("library-tab-attended").click();
  await expect(page.getByTestId(`library-row-${id}`)).toBeVisible();
});

// The festival path is a different shape: a lineup pasted as text becomes performances,
// and the detail page reads them back.
test("a festival keeps the lineup pasted into the form", async ({ page, request }) => {
  const t = await token(request);
  const name = `Dolina Fest ${uniq()}`;

  await signIn(page, request);
  await page.goto("/library");
  await page.getByTestId("library-add-event").click();

  await page.getByTestId("manual-name").fill(name);
  await page.getByTestId("manual-start-date").fill(daysFromNow(-70));
  await optionSelect(page, "manual-kind-festival").selectOption("festival");

  await page.getByTestId("lineup-paste").click();
  await page.getByTestId("lineup-csv").fill("First Act\nSecond Act\nThird Act");
  await page.getByTestId("lineup-csv-apply").click();
  await expect(page.getByTestId("lineup-act-0")).toBeVisible();
  await expect(page.getByTestId("lineup-act-2")).toBeVisible();

  await optionSelect(page, "manual-status-attended").selectOption("attended");
  await page.getByTestId("manual-save").click();

  const id = await eventIdByName(request, t, name);
  await page.goto(`/events/${id}`);
  await expect(page.getByTestId("event-detail-lineup")).toContainText("Second Act");
});
