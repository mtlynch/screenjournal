import { expect } from "@playwright/test";

export async function registerAsAdmin(page) {
  await page.goto("/");

  await page.locator("data-test-id=sign-up-btn").click();

  await expect(page).toHaveURL("/sign-up");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=email").fill("admin@example.com");
  await page.locator("id=password").fill("dummypass");
  await page.locator("id=password-confirm").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
}
