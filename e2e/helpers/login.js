import { expect } from "@playwright/test";

export async function loginAsUser(page, username, password) {
  await page.goto("/");

  await page.getByRole("menuitem", { name: "Log in" }).click();

  await page.getByRole("textbox", { name: /username/i }).fill(username);
  await page.getByRole("textbox", { name: /password/i }).fill(password);
  await page.getByRole("button", { name: "Log in" }).click();

  await expect(page).toHaveURL("/reviews");
}

export async function loginAsAdmin(page) {
  await loginAsUser(page, "dummyadmin", "dummypass");
}

export async function loginAsUserA(page) {
  await loginAsUser(page, "userA", "password123");
}

export async function loginAsUserB(page) {
  await loginAsUser(page, "userB", "password456");
}
