import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db.js";
import { loginAsUser, loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("user can change their password and log in again", async ({ page }) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Security" }).click();

  await expect(page).toHaveURL("/account/security");
  await page.getByText("Change password").click();

  await expect(page).toHaveURL("/account/change-password");
  await page.getByLabel("Current Password").fill("password123");
  await page.getByLabel(/^New Password$/).fill("password321");
  await page.getByLabel("Confirm New Password").fill("password321");
  await page.getByRole("button", { name: /Change password/i }).click();

  await expect(page).toHaveURL("/account/security");

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();

  await loginAsUser(page, "userA", "password321");
});

test("user can cancel their password password change and log in with their original credentials", async ({
  page,
}) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Security" }).click();

  await expect(page).toHaveURL("/account/security");
  await page.getByText("Change password").click();

  await expect(page).toHaveURL("/account/change-password");
  await page.getByLabel("Current Password").fill("password123");
  await page.getByLabel(/^New Password$/).fill("password321");
  await page.getByLabel("Confirm New Password").fill("password321");
  await page.getByRole("button", { name: /Cancel/i }).click();

  await expect(page).toHaveURL("/account/security");

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();

  await loginAsUserA(page);
});

test("user can't change their password if their original is incorrect", async ({
  page,
}) => {
  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Security" }).click();

  await expect(page).toHaveURL("/account/security");
  await page.getByText("Change password").click();

  await expect(page).toHaveURL("/account/change-password");
  await page.getByLabel("Current Password").fill("imawrongpassword");
  await page.getByLabel(/^New Password$/).fill("password321");
  await page.getByLabel("Confirm New Password").fill("password321");
  await page.getByRole("button", { name: /Change password/i }).click();

  await expect(page).toHaveURL("/account/change-password");

  await expect(page.getByText("Incorrect password")).toBeVisible();
});
