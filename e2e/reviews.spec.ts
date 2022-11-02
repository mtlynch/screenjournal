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
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-solid")
  ).toHaveCount(3);
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-regular")
  ).toHaveCount(7);
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
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-solid")
  ).toHaveCount(10);
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
  await expect(reviewCard.locator(".card-footer")).toHaveText(
    "Watched 2022-10-29"
  );
});

test("adds a new rating and edits the details", async ({ page }) => {
  await login(page);

  await page.locator("data-test-id=add-rating").click();

  await page.locator("#media-title").fill("There's Something About Mary");

  await page.locator("#rating-select").selectOption({ label: "10" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.locator("data-test-id=edit-rating").click;

  await expect(page.locator("h1")).toHaveText("There's Something About Mary");

  await expect(reviewCard.locator(".card-title")).toHaveText(
    "There's Something About Mary"
  );

  await page.locator("#rating-select").selectOption({ label: "8" });

  await page.locator("#watched-date").fill("2022-10-22");
  await page.locator("#blurb").fill("Not as good as I remembered");

  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-solid")
  ).toHaveCount(8);
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-regular")
  ).toHaveCount(2);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Not as good as I remembered"
  );
  await expect(reviewCard.locator(".card-footer")).toHaveText(
    "Watched 2022-10-29"
  );
});

test("adds a new rating and cancels the edit", async ({ page }) => {
  await login(page);

  await page.locator("data-test-id=add-rating").click();

  await page.locator("#media-title").fill("The Matrix");
  await page.locator("#rating-select").selectOption({ label: "8" });
  await page.locator("#watched-date").fill("2022-10-29");
  await page.locator("#blurb").fill("Am I in the Matrix right now?");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.locator("data-test-id=edit-rating").click;

  // Make edits that will be ignored when we cancel the edit.
  await page.locator("#rating-select").selectOption({ label: "1" });
  await page.locator("#watched-date").fill("2022-10-22");
  await page.locator("#blurb").fill("Ignore this edit");

  await reviewCard.locator("data-test-id=cancel-edit").click();

  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-solid")
  ).toHaveCount(8);
  await expect(
    reviewCard.locator(".card-subtitle .fa-star.fa-regular")
  ).toHaveCount(2);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Am I in the Matrix right now?"
  );
  await expect(reviewCard.locator(".card-footer")).toHaveText(
    "Watched 2022-10-29"
  );
});
