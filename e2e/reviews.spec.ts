import { test, expect } from "@playwright/test";
import { login } from "./helpers/login.js";
import { wipeDB } from "./helpers/wipe.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await login(page);
});

test("adds a new rating and fills in only required fields", async ({
  page,
}) => {
  await page.locator("data-test-id=add-rating").click();

  await page.locator("title-search #media-title").fill("slow lear");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("Slow Learners (2015)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "Slow Learners"
  );

  await page.locator("#rating-select").selectOption({ label: "3" });

  await page.locator("#watched-date").fill("2022-10-21");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText("Slow Learners");
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /dummyuser watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-test-id='watch-date']")
  ).toHaveAttribute("title", "2022-10-21");
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(3);
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(7);
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
});

test("adds a new rating and fills all fields", async ({ page }) => {
  await page.locator("data-test-id=add-rating").click();

  await page.locator("title-search #media-title").fill("eternal sunshine");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText(
    "Eternal Sunshine of the Spotless Mind (2004)"
  );
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "Eternal Sunshine of the Spotless Mind"
  );

  await page.locator("#rating-select").selectOption({ label: "10" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText(
    "Eternal Sunshine of the Spotless Mind"
  );
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /dummyuser watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-test-id='watch-date']")
  ).toHaveAttribute("title", "2022-10-29");
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(10);
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
});

test("adds a new rating and edits the details", async ({ page }) => {
  await page.locator("data-test-id=add-rating").click();

  await page.locator("title-search #media-title").fill("something about ma");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("There's Something About Mary (1998)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "There's Something About Mary"
  );

  await page.locator("#rating-select").selectOption({ label: "10" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.locator("data-test-id=edit-rating").click();

  await expect(page.locator("h1")).toHaveText(
    "There's Something About Mary (1998)"
  );

  await expect(page.locator("#rating-select")).toHaveValue("10");
  await page.locator("#rating-select").selectOption({ label: "8" });

  await expect(page.locator("#watched-date")).toHaveValue("2022-10-29");
  await page.locator("#watched-date").fill("2022-10-22");

  await expect(page.locator("#blurb")).toHaveValue("My favorite movie!");
  await page.locator("#blurb").fill("Not as good as I remembered");

  await page.locator("form .btn-primary").click();

  await expect(page).toHaveURL("/reviews");

  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /dummyuser watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-test-id='watch-date']")
  ).toHaveAttribute("title", "2022-10-22");
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(8);
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(2);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Not as good as I remembered"
  );
});

test("adds a new rating and cancels the edit", async ({ page }) => {
  await page.locator("data-test-id=add-rating").click();

  await page.locator("title-search #media-title").fill("the matri");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("The Matrix (1999)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "The Matrix"
  );

  await page.locator("#rating-select").selectOption({ label: "8" });
  await page.locator("#watched-date").fill("2022-10-29");
  await page.locator("#blurb").fill("Am I in the Matrix right now?");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.locator("data-test-id=edit-rating").click();

  // Make edits that will be ignored when we cancel the edit.
  await page.locator("#rating-select").selectOption({ label: "1" });
  await page.locator("#watched-date").fill("2022-10-22");
  await page.locator("#blurb").fill("Ignore this edit");

  await page.locator("data-test-id=cancel-edit").click();

  await expect(page).toHaveURL("/reviews");

  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /dummyuser watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-test-id='watch-date']")
  ).toHaveAttribute("title", "2022-10-29");
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(8);
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(2);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Am I in the Matrix right now?"
  );
});
