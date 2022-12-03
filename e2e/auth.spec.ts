import { test, expect } from "@playwright/test";

test("logs in and logs out", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-in-btn").click();

  await expect(page).toHaveURL("/login");
  await page.locator("id=username").fill("dummyuser");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.locator("#account-dropdown").click();
  await page.locator("#navbar-log-out").click();

  await expect(page).toHaveURL("/");
  await expect(page.locator("data-test-id=sign-in-btn")).toBeVisible();
});
