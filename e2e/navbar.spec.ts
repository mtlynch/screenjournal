import { test, expect } from "@playwright/test";

test("navbar updates links based on auth state", async ({ page }) => {
  await page.goto("/");

  await page.locator(".navbar-brand").click();
  await expect(page).toHaveURL("/");

  await page.locator(".navbar").getByText("Home").click();
  await expect(page).toHaveURL("/");

  await page.locator(".navbar").getByText("About").click();
  await expect(page).toHaveURL("/about");

  await expect(page.locator(".navbar").getByText("Account")).toHaveCount(0);

  await page.goto("/login");

  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  await page.locator(".navbar-brand").click();
  await expect(page).toHaveURL("/reviews");

  await page.locator(".navbar").getByText("Home").click();
  await expect(page).toHaveURL("/reviews");

  await page.locator(".navbar").getByText("About").click();
  await expect(page).toHaveURL("/about");

  await expect(page.locator(".navbar").getByText("Account")).toHaveCount(1);
});
