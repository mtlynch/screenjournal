import { expect, Page, Cookie } from "@playwright/test";

export async function populateDummyData(page: Page): Promise<void> {
  await page.goto("/"); // hack to populate the DB cookie
  const response = await page.goto("/api/debug/db/populate-dummy-data");
  await expect(response?.status()).toBe(200);
  await page.goto("/");
}

export function readDbTokenCookie(cookies: Cookie[]): Cookie {
  for (const cookie of cookies) {
    if (cookie.name === "db-token") {
      return cookie;
    }
  }
  throw new Error("db-token cookie not found");
}
