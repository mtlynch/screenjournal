import { test, expect } from "@playwright/test";
import { login } from "./helpers/login.js";

test("adds a new rating and fills all fields", async ({ page }) => {
  await login(page);

  await page.locator("data-test-id=add-rating").click();

  await page
    .locator("#media-title")
    .fill("Eternal Sunshine of the Spotless Mind");

  await page.locator("#rating-select").selectOption({ label: "10" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
});
