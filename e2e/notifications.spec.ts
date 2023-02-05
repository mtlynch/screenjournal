import { test, expect } from "@playwright/test";
import { populateDummyData, wipeDB } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("TODO", async ({ page }) => {
  await page.locator("#account-dropdown").click();
  await page.locator("data-test-id=notification-prefs-btn").click();

  await expect(page).toHaveURL("/account/notifications");

  //TODO
});
