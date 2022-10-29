import { test, expect } from "@playwright/test";

test("logs in", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-in-btn").click();

  await expect(page).toHaveURL("/login");
  await page.locator("form input[type='username']").fill("dummyuser");
  await page.locator("form input[type='password']").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
});
