import { test, expect } from "@playwright/test";
import { loginAsAdmin } from "./helpers/login";
import { populateDummyData, readDbTokenCookie } from "./helpers/db";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsAdmin(page);
});

test("admin can create password reset link and user can reset password", async ({
  page,
  browser,
}) => {
  // Admin creates reset link for userA
  await page.getByRole("menuitem", { name: "Admin" }).click();
  await page.getByRole("menuitem", { name: "Password Resets" }).click();

  await expect(page).toHaveURL("/admin/reset-password");
  await page.selectOption("#username", "userA");
  await page.getByRole("button", { name: "Generate Reset Link" }).click();

  // Wait for the reset link to be generated and extract the token
  await expect(page.locator("[data-token]")).toBeVisible();
  const token = await page
    .locator("[data-token]")
    .first()
    .getAttribute("data-token");
  expect(token).toBeTruthy();

  const resetLink = `/account/password-reset?token=${token}`;

  // Switch to user context
  const userContext = await browser.newContext();

  // Share database across users
  await userContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);

  const userPage = await userContext.newPage();
  await userPage.goto(resetLink);

  await expect(userPage).toHaveURL(resetLink);

  // Reset password (no current password field should be visible)
  await expect(userPage.getByLabel("Current Password")).not.toBeVisible();
  await userPage.getByLabel(/^New Password$/).fill("newpassword123");
  await userPage.getByLabel("Confirm New Password").fill("newpassword123");
  await userPage.getByRole("button", { name: /Reset password/i }).click();

  // Verify successful reset
  await expect(userPage.getByText("Password updated")).toBeVisible();

  // User should be redirected to home after password reset
  await expect(userPage).toHaveURL("/reviews");

  // Navigate to "My ratings" to verify we're logged in as the correct user
  await userPage.getByRole("menuitem", { name: "Account" }).click();
  await userPage.getByRole("menuitem", { name: "My ratings" }).click();

  // Should be on userA's ratings page
  await expect(userPage).toHaveURL("/reviews/by/userA");

  await userContext.close();
});

test("password reset link expires properly", async ({ page, browser }) => {
  // Admin creates reset link for userA.
  await page.getByRole("menuitem", { name: "Admin" }).click();
  await page.getByRole("menuitem", { name: "Password Resets" }).click();

  await expect(page).toHaveURL("/admin/reset-password");
  await page.selectOption("#username", "userA");
  await page.getByRole("button", { name: "Generate Reset Link" }).click();

  // Wait for the reset link to be generated and extract the token.
  await expect(page.locator("[data-token]")).toBeVisible();
  const token = await page
    .locator("[data-token]")
    .first()
    .getAttribute("data-token");
  expect(token).toBeTruthy();

  // Confirm the dialog to confirm delete.
  page.on("dialog", (dialog) => dialog.accept());

  // Admin deletes the reset token.
  await page
    .locator("[data-token]")
    .first()
    .locator("..")
    .getByRole("button", { name: /Delete/i })
    .click();

  // Wait for the row to disappear from the table.
  await expect(page.locator("[data-token]")).not.toBeVisible();

  // Switch to user context
  const userContext = await browser.newContext();
  await userContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);

  const userPage = await userContext.newPage();
  const resetLink = `/account/password-reset?token=${token}`;

  // Trying to access expired/deleted token should fail.
  const response = await userPage.goto(resetLink);
  await expect(response?.status()).toBe(401);

  await userContext.close();
});
