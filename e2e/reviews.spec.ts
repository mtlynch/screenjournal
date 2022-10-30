import { test, expect } from "@playwright/test";
import { login } from "./helpers/login.js";

test("adds a new rating and fills in only required fields", async ({
  page,
}) => {
  await login(page);

  await page.locator("data-test-id=add-rating").click();

  await page.locator("#media-title").fill("Slow Learners");

  await page.locator("#rating-select").selectOption({ label: "3" });

  await page.locator("#watched-date").fill("2022-10-21");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText("Slow Learners");
  await expect(reviewCard.locator(".card-subtitle")).toHaveText("3 / 10");
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
  await expect(reviewCard.locator(".card-footer")).toHaveText(
    "Watched 2022-10-21"
  );
});

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

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText(
    "Eternal Sunshine of the Spotless Mind"
  );
  await expect(reviewCard.locator(".card-subtitle")).toHaveText("10 / 10");
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
  await expect(reviewCard.locator(".card-footer")).toHaveText(
    "Watched 2022-10-29"
  );
});
