import { test, expect } from "@playwright/test";
import { registerAsAdmin } from "./helpers/auth.js";
import { wipeDB } from "./helpers/wipe.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await registerAsAdmin(page);
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
    /dummyadmin watched this .+ ago/,
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

test("adds a new rating that's too long to display in a card", async ({
  page,
}) => {
  await page.locator("data-test-id=add-rating").click();

  await page
    .locator("title-search #media-title")
    .fill("Weird: The Al Yankovic Story");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("Weird: The Al Yankovic Story (2022)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "Weird: The Al Yankovic Story"
  );
  await page.locator("#watched-date").fill("2022-11-11");

  await page.locator("#rating-select").selectOption({ label: "10" });

  await page.locator("#blurb")
    .fill(`Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.

Daniel Radcliffe is fantastic, and it's a great film role for Rainn Wilson. There are a million great cameos.

You'll like it if you enjoy things like Children's Hospital, Comedy Bang Bang, or Popstar.`);

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText(
    "Weird: The Al Yankovic Story"
  );
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /dummyadmin watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-test-id='watch-date']")
  ).toHaveAttribute("title", "2022-11-11");
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(10);
  await expect(
    reviewCard.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);

  await expect(reviewCard.locator(".card-text")).toHaveText(
    `Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always e...`
  );
  await reviewCard.locator("data-test-id=full-review").click();

  await expect(page.locator("h1")).toHaveText(
    "Weird: The Al Yankovic Story (2022)"
  );
  await expect(page.locator(".card-subtitle")).toHaveText(
    /dummyadmin watched this .+ ago/,
    { useInnerText: true }
  );

  await expect(
    page.locator("[data-test-id='rating'] .fa-star.fa-solid")
  ).toHaveCount(10);
  await expect(
    page.locator("[data-test-id='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);

  await expect(page.locator("data-test-id=blurb")).toHaveText(
    `Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.

Daniel Radcliffe is fantastic, and it's a great film role for Rainn Wilson. There are a million great cameos.

You'll like it if you enjoy things like Children's Hospital, Comedy Bang Bang, or Popstar.`
  );
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
    /dummyadmin watched this .+ ago/,
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
    /dummyadmin watched this .+ ago/,
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
    /dummyadmin watched this .+ ago/,
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
