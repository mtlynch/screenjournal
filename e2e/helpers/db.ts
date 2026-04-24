import { expect, Page, Cookie } from "@playwright/test";

export async function populateDummyData(page: Page): Promise<void> {
  // Navigate to root to get the db-token cookie
  await page.goto("/");

  // Verify we got the db-token cookie
  const cookies = await page.context().cookies();
  const dbToken = cookies.find((c) => c.name === "db-token");
  if (!dbToken) {
    throw new Error("populateDummyData: db-token cookie not set after goto /");
  }

  // Populate the test database
  const response = await page.goto("/api/debug/db/populate-dummy-data");
  await expect(response?.status()).toBe(200);

  // Wait for database to be fully initialized
  await page.waitForTimeout(200);

  // Navigate back to root
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
