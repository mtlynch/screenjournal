import { expect, Page } from "@playwright/test";

export async function populateDummyData(page: Page): Promise<void> {
  const response = await page.goto("/api/debug/db/populate-dummy-data");
  await expect(response?.status()).toBe(200);
  await page.goto("/");
}
