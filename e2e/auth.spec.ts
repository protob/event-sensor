import { test, expect } from "@playwright/test";

test("a protected route sends an anonymous visitor to login and back", async ({ page }) => {
  await page.goto("/library");
  await expect(page).toHaveURL(/\/login\?redirect=/);

  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("password");
  await page.getByTestId("login-submit").click();

  await expect(page).toHaveURL(/\/library$/);
  await expect(page.getByTestId("library-tab-interested")).toBeVisible();
});

test("a wrong password reports and stays put", async ({ page }) => {
  await page.goto("/login");
  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("not-the-password");
  await page.getByTestId("login-submit").click();

  await expect(page.getByTestId("login-error")).toBeVisible();
  await expect(page).toHaveURL(/\/login$/);
});

// The session is opened through the form here rather than through the helper, which
// re-seeds the token on every navigation and would undo the logout under test.
test("logging out drops the session", async ({ page }) => {
  await page.goto("/library");
  await page.getByTestId("login-username").fill("admin");
  await page.getByTestId("login-password").fill("password");
  await page.getByTestId("login-submit").click();
  await expect(page.getByTestId("user-menu")).toBeVisible();

  await page.getByTestId("user-menu").click();
  await page.getByTestId("logout").click();

  await page.goto("/library");
  await expect(page).toHaveURL(/\/login/);
});
