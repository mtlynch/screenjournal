import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsUserA } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("notifications page reflects the backend store for new reviews", async ({
  page,
}) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Notifications" }).click();

  await expect(page).toHaveURL("/account/notifications");

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).toBeChecked();

  // Turn off new review notifications.
  await page.getByLabel("Email me when users post reviews").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).not.toBeChecked();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).not.toBeChecked();

  // Turn on new review notifications.
  await page.getByLabel("Email me when users post reviews").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).toBeChecked();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).toBeChecked();
});

test("notifications page reflects the backend store for new comments", async ({
  page,
}) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Notifications" }).click();

  await expect(page).toHaveURL("/account/notifications");

  await expect(
    page.getByLabel("Email me when someone replies to me in a review comment")
  ).toBeChecked();

  // Turn off new comment notifications.
  await page
    .getByLabel("Email me when someone replies to me in a review comment")
    .click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when someone replies to me in a review comment")
  ).not.toBeChecked();

  await page.reload();

  await expect(
    page.getByLabel("Email me when someone replies to me in a review comment")
  ).not.toBeChecked();

  // Turn on new comment notifications.
  await page
    .getByLabel("Email me when someone replies to me in a review comment")
    .click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when someone replies to me in a review comment")
  ).toBeChecked();

  await page.reload();

  await expect(
    page.getByLabel("Email me when someone replies to me in a review comment")
  ).toBeChecked();
});
