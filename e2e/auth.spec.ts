import { test, expect } from "@playwright/test";
import { wipeDB } from "./helpers/db.js";

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

test("sign up fails if passwords are different", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("dummypass");
  await page.locator("id=password-confirm").fill("otherpass");

  await expect(page.locator("#password-confirm:invalid")).toBeFocused();

  // Submitting form should fail.
  await page.locator("form input[type='submit']").click();
  await expect(page).toHaveURL("/sign-up");
});

test("sign up fails if username is invalid", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("root");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("password");
  await page.locator("id=password-confirm").fill("password");
  await page.locator("form input[type='submit']").click();

  await expect(page.locator("id=error")).toBeVisible();

  await expect(page).toHaveURL("/sign-up");
});

test("sign up fails if password is too short", async ({ page }) => {
  await page.goto("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("pass");
  await page.locator("id=password-confirm").fill("pass");
  await page.locator("form input[type='submit']").click();

  await expect(page.locator("#password:invalid")).toBeFocused();

  // Submitting form should fail.
  await page.locator("form input[type='submit']").click();
  await expect(page).toHaveURL("/sign-up");
});

test("signs up fails after there's already an admin user", async ({ page }) => {
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

  // Attempt to sign up again
  await expect(page).toHaveURL("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");

  await expect(page.locator("form input[type='submit']")).not.toBeVisible();
});
