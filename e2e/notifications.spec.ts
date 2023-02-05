import { test, expect } from "@playwright/test";
import { populateDummyData, wipeDB } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("notifications page reflects the backend store", async ({ page }) => {
  await page.locator("#account-dropdown").click();
  await page.locator("data-test-id=notification-prefs-btn").click();

  await expect(page).toHaveURL("/account/notifications");

  await expect(page.locator("#new-reviews-checkbox")).toBeChecked();
  await expect(page.locator("#notifications-form .btn-primary")).toBeDisabled();

  // Turn off new review notifications
  await page.locator("#new-reviews-checkbox").click();
  await page.locator("#notifications-form .btn-primary").click();

  await expect(page.locator("#new-reviews-checkbox")).not.toBeChecked();
  await expect(page.locator("#notifications-form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(page.locator("#new-reviews-checkbox")).not.toBeChecked();
  await expect(page.locator("#notifications-form .btn-primary")).toBeDisabled();

  // Turn on new review notifications
  await page.locator("#new-reviews-checkbox").click();
  await page.locator("#notifications-form .btn-primary").click();

  await expect(page.locator("#new-reviews-checkbox")).toBeChecked();
  await expect(page.locator("#notifications-form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(page.locator("#new-reviews-checkbox")).toBeChecked();
  await expect(page.locator("#notifications-form .btn-primary")).toBeDisabled();
});
