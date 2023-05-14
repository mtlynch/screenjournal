import { expect } from "@playwright/test";

export async function populateDummyData(page) {
  await page.goto("/"); // hack to populate the DB cookie
  const response = await page.goto("/api/debug/db/populate-dummy-data");
  await expect(response?.status()).toBe(200);
}

export function readDbTokenCookie(cookies) {
  for (const cookie of cookies) {
    if (cookie.name === "db-token") {
      return cookie;
    }
  }
  return undefined;
}
