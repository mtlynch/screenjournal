import { expect } from "@playwright/test";

export async function login(page) {
  await page.goto("/");

  await page.locator("data-test-id=sign-in-btn").click();

  await expect(page).toHaveURL("/login");
  await page.locator("id=username").fill("dummyuser");
  await page.locator("id=password").fill("dummypass");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
}
