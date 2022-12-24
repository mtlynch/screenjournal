import { test, expect } from "@playwright/test";
import { wipeDB } from "./helpers/wipe.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
});

test("signs up and logs out and signs in again", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("dummypass");
  await page.locator("id=password-confirm").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.locator("#account-dropdown").click();
  await page.locator("#navbar-log-out").click();

  // Sign in again
  await expect(page).toHaveURL("/");
  await page.locator("data-test-id=sign-in-btn").click();

  await expect(page).toHaveURL("/login");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.locator("#account-dropdown").click();
  await page.locator("#navbar-log-out").click();

  await expect(page).toHaveURL("/");
  await expect(page.locator("data-test-id=sign-in-btn")).toBeVisible();
});
