import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsUserA } from "./helpers/login";

test("signs up and logs out and signs in again", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("menuitem", { name: "Sign up" }).click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("dummypass");
  await page.locator("id=password-confirm").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();

  // Sign in again
  await expect(page).toHaveURL("/");
  await page.getByRole("menuitem", { name: "Log in" }).click();

  await expect(page).toHaveURL("/login");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();

  await expect(page).toHaveURL("/");
  await expect(page.getByRole("menuitem", { name: "Log in" })).toBeVisible();
});

test("sign up fails if passwords are different", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("menuitem", { name: "Sign up" }).click();

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

  await page.getByRole("menuitem", { name: "Sign up" }).click();

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

  await page.getByRole("menuitem", { name: "Sign up" }).click();

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

  await page.getByRole("menuitem", { name: "Sign up" }).click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("dummypass");
  await page.locator("id=password-confirm").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();

  // Attempt to sign up again
  await expect(page).toHaveURL("/");

  await page.getByRole("menuitem", { name: "Sign up" }).click();

  await expect(page).toHaveURL("/sign-up");

  await expect(page.locator("form input[type='submit']")).not.toBeVisible();
});

test("prompts to log in if they click a link to a review before signing in", async ({
  page,
}) => {
  await populateDummyData(page);

  await page.goto("/movies/1");

  await expect(page).toHaveURL("/login?next=%2Fmovies%2F1");
  await page.locator("id=username").fill("userB");
  await page.locator("id=password").fill("password456");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/1");
});

test("root route redirects a logged in user to the reviews index", async ({
  page,
}) => {
  await populateDummyData(page);
  await loginAsUserA(page);
  await page.goto("/");
  await expect(page).toHaveURL("/reviews");
});
