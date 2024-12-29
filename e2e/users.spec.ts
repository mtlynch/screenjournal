import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

async function normalizedTextContent(locator: Locator): Promise<string> {
  const text = await locator.textContent();
  return text?.replace(/\s+/g, " ").trim() ?? "";
}

test("views list of users", async ({ page }) => {
  await page.getByRole("menuitem", { name: "Users" }).click();

  await expect(page).toHaveURL("/users");

  expect(await normalizedTextContent(page.locator("ol li").nth(0))).toMatch(
    /dummyadmin joined \d{4}-\d{2}-\d{2} and has written 0 reviews\./
  );
  expect(await normalizedTextContent(page.locator("ol li").nth(1))).toMatch(
    /userA joined \d{4}-\d{2}-\d{2} and has written 1 reviews\./
  );
  expect(await normalizedTextContent(page.locator("ol li").nth(2))).toMatch(
    /userB joined \d{4}-\d{2}-\d{2} and has written 3 reviews\./
  );
});
