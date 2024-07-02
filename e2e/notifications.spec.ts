import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

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
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  // Turn off new review notifications.
  await page.getByLabel("Email me when users post reviews").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).not.toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).not.toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  // Turn on new review notifications.
  await page.getByLabel("Email me when users post reviews").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users post reviews")
  ).toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();
});

test("notifications page reflects the backend store for new comments", async ({
  page,
}) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Notifications" }).click();

  await expect(page).toHaveURL("/account/notifications");

  await expect(
    page.getByLabel("Email me when users add comments")
  ).toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  // Turn off new comment notifications.
  await page.getByLabel("Email me when users add comments").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users add comments")
  ).not.toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users add comments")
  ).not.toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  // Turn on new comment notifications.
  await page.getByLabel("Email me when users add comments").click();
  await page.locator("form .btn-primary").click();

  await expect(
    page.getByLabel("Email me when users add comments")
  ).toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();

  await page.reload();

  await expect(
    page.getByLabel("Email me when users add comments")
  ).toBeChecked();
  await expect(page.locator("form .btn-primary")).toBeDisabled();
});
