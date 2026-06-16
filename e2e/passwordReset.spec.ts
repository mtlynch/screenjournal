import { test, expect } from "./fixtures";
import type { Page } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsUser } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
});

async function resetFormIsAvailable(page: Page): Promise<boolean> {
  await page.goto("/reset-password");
  const resetForm = page.locator("#reset-password-form");
  if (await resetForm.isVisible()) {
    return true;
  }

  await expect(page.getByRole("alert")).toContainText(
    "Password resets are not available on this server."
  );
  return false;
}

test("user can reset password via forgot-password flow", async ({
  page,
  browser,
  baseURL,
}) => {
  if (!(await resetFormIsAvailable(page))) {
    return;
  }

  // Submit the email.
  await page.locator("id=email").fill("userA@example.com");
  await page.getByRole("button", { name: "Send reset link" }).click();

  // Verify success message.
  await expect(page.getByText("Request successful!")).toBeVisible();

  // Retrieve the token from the debug endpoint.
  const apiResponse = await page.request.get(
    "/api/debug/password-reset-token/userA"
  );
  expect(apiResponse.status()).toBe(200);
  const { token } = await apiResponse.json();
  expect(token).toBeTruthy();

  const resetLink = `/account/password-reset?username=userA&token=${token}`;

  // Switch to a fresh user context.
  const userContext = await browser.newContext();
  const userPage = await userContext.newPage();

  await userPage.goto(`${baseURL}${resetLink}`);
  await expect(userPage).toHaveURL(`${baseURL}${resetLink}`);

  await userPage.route(
    /\/account\/password-reset\?username=userA&token=.*/,
    async (route) => {
      if (route.request().method() === "PUT") {
        await new Promise((resolve) => setTimeout(resolve, 500));
      }
      await route.continue();
    }
  );

  // Reset password.
  await expect(userPage.getByLabel("Current Password")).not.toBeVisible();
  await userPage.getByLabel(/^New Password$/).fill("brandnewpass99");
  await userPage.getByLabel("Confirm New Password").fill("brandnewpass99");
  const passwordResetResponse = userPage.waitForResponse((response) => {
    return (
      response.request().method() === "PUT" &&
      response.url().includes("/account/password-reset?username=userA&token=")
    );
  });
  await userPage.getByRole("button", { name: /Reset password/i }).click();
  await expect(userPage.getByLabel(/^New Password$/)).toBeDisabled();
  await expect(userPage.getByLabel("Confirm New Password")).toBeDisabled();
  await passwordResetResponse;

  // Verify successful reset.
  await expect(userPage.getByText("Password updated")).toBeVisible();
  await expect(userPage.locator("#password-form")).toBeHidden();

  // User should be redirected to reviews after password reset.
  await expect(userPage).toHaveURL(`${baseURL}/reviews`);

  // Verify logged in as the correct user.
  await userPage.getByRole("menuitem", { name: "Account" }).click();
  await userPage.getByRole("menuitem", { name: "My ratings" }).click();
  await expect(userPage).toHaveURL(`${baseURL}/reviews/by/userA`);

  await userContext.close();

  // Verify the new password works by logging in.
  await loginAsUser(page, "userA", "brandnewpass99");
});

test("reset shows same success message for unknown email", async ({ page }) => {
  if (!(await resetFormIsAvailable(page))) {
    return;
  }

  await page.locator("id=email").fill("unknown@example.com");
  await page.getByRole("button", { name: "Send reset link" }).click();

  // Should show the same success message to prevent email enumeration.
  await expect(page.getByText("Request successful!")).toBeVisible();
});

test("incorrect password reset token is rejected", async ({
  browser,
  baseURL,
}) => {
  const bogusToken = "ABCDEFGHJKLMNPQRSTUVWXYZabcdef23";
  const resetLink = `/account/password-reset?username=userA&token=${bogusToken}`;

  // Switch to a fresh user context.
  const userContext = await browser.newContext();
  const userPage = await userContext.newPage();

  // Navigate to the reset page with a bogus token.
  const response = await userPage.goto(`${baseURL}${resetLink}`);
  expect(response?.status()).toBe(401);

  await userContext.close();
});
