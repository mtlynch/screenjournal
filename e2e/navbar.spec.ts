import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsAdmin } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
});

test("navbar updates links based on auth state", async ({ page }) => {
  await page.goto("/");

  await page.locator(".navbar-brand").click();
  await expect(page).toHaveURL("/");

  await page.locator(".navbar").getByText("Home").click();
  await expect(page).toHaveURL("/");

  await page.locator(".navbar").getByText("About").click();
  await expect(page).toHaveURL("/about");

  await expect(page.locator(".navbar").getByText("Account")).toHaveCount(0);

  await loginAsAdmin(page);

  await page.locator(".navbar-brand").click();
  await expect(page).toHaveURL("/reviews");

  await page.locator(".navbar").getByText("Home").click();
  await expect(page).toHaveURL("/reviews");

  await page.locator(".navbar").getByText("About").click();
  await expect(page).toHaveURL("/about");

  await expect(page.locator(".navbar").getByText("Account")).toHaveCount(1);
});
