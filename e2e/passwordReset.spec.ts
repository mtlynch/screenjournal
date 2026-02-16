import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db";
import { loginAsUser } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
});

test("user can reset password via forgot-password flow", async ({
  page,
  browser,
}) => {
  // Navigate to reset password page from login.
  await page.getByRole("menuitem", { name: "Log in" }).click();
  await page.getByRole("link", { name: "Forgot password?" }).click();

  await expect(page).toHaveURL("/reset-password");

  // Submit the email.
  await page.locator("id=email").fill("userA@example.com");
  await page.getByRole("button", { name: "Send reset link" }).click();

  // Verify success message.
  await expect(page.getByText("Reset successful!")).toBeVisible();

  // Retrieve the token from the debug endpoint.
  const dbCookie = readDbTokenCookie(await page.context().cookies());
  const apiResponse = await page.request.get(
    "/api/debug/password-reset-token/userA",
    {
      headers: {
        Cookie: `${dbCookie.name}=${dbCookie.value}`,
      },
    }
  );
  expect(apiResponse.status()).toBe(200);
  const { token } = await apiResponse.json();
  expect(token).toBeTruthy();

  const resetLink = `/account/password-reset?username=userA&token=${token}`;

  // Switch to a fresh user context.
  const userContext = await browser.newContext();
  await userContext.addCookies([dbCookie]);
  const userPage = await userContext.newPage();

  await userPage.goto(resetLink);
  await expect(userPage).toHaveURL(resetLink);

  // Reset password.
  await expect(userPage.getByLabel("Current Password")).not.toBeVisible();
  await userPage.getByLabel(/^New Password$/).fill("brandnewpass99");
  await userPage.getByLabel("Confirm New Password").fill("brandnewpass99");
  await userPage.getByRole("button", { name: /Reset password/i }).click();

  // Verify successful reset.
  await expect(userPage.getByText("Password updated")).toBeVisible();

  // User should be redirected to reviews after password reset.
  await expect(userPage).toHaveURL("/reviews");

  // Verify logged in as the correct user.
  await userPage.getByRole("menuitem", { name: "Account" }).click();
  await userPage.getByRole("menuitem", { name: "My ratings" }).click();
  await expect(userPage).toHaveURL("/reviews/by/userA");

  await userContext.close();

  // Verify the new password works by logging in.
  await loginAsUser(page, "userA", "brandnewpass99");
});

test("reset shows same success message for unknown email", async ({ page }) => {
  await page.goto("/reset-password");

  await page.locator("id=email").fill("unknown@example.com");
  await page.getByRole("button", { name: "Send reset link" }).click();

  // Should show the same success message to prevent email enumeration.
  await expect(page.getByText("Reset successful!")).toBeVisible();
});

test("incorrect password reset token is rejected", async ({
  page,
  browser,
}) => {
  const bogusToken = "ABCDEFGHJKLMNPQRSTUVWXYZabcdef23";
  const resetLink = `/account/password-reset?username=userA&token=${bogusToken}`;

  // Switch to a fresh user context.
  const userContext = await browser.newContext();
  await userContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const userPage = await userContext.newPage();

  // Navigate to the reset page with a bogus token.
  const response = await userPage.goto(resetLink);
  expect(response?.status()).toBe(401);

  await userContext.close();
});
