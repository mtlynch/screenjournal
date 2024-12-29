import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("views list of users", async ({ page }) => {
  await page.getByRole("menuitem", { name: "Users" }).click();

  await expect(page).toHaveURL("/users");

  await expect(page.locator("ol li:nth-child(2)")).toHaveText(
    /dummyadmin joined \d{4}-\d{2}-\d{2} and has written 0 reviews\./
  );
  await expect(page.locator("ol li:nth-child(2)")).toHaveText(
    /userA joined \d{4}-\d{2}-\d{2} and has written 1 reviews\./
  );
  await expect(page.locator("ol li:nth-child(3)")).toHaveText(
    /userB joined \d{4}-\d{2}-\d{2} and has written 3 reviews\./
  );
});
