import { expect } from "@playwright/test";

export async function loginAsAdmin(page) {
  await page.goto("/");

  await page.locator("data-test-id=sign-in-btn").click();

  await expect(page).toHaveURL("/login");
  await page.locator("id=username").fill("dummyadmin");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
}
