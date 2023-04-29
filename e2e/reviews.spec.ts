import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";
import { loginAsUserA, loginAsUserB } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("adds a new rating and fills in only required fields", async ({
  page,
}) => {
  await page.getByTestId("add-rating").click();

  await page.locator("title-search #media-title").fill("slow lear");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("Slow Learners (2015)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "Slow Learners"
  );

  await page.locator("#rating-select").selectOption({ label: "1" });

  await page.locator("#watched-date").fill("2022-10-21");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText("Slow Learners");
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2022-10-21");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(4);
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
});

test("adds a new rating that's too long to display in a card", async ({
  page,
}) => {
  await page.getByTestId("add-rating").click();

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

  await page.locator("#rating-select").selectOption({ label: "5" });

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
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2022-11-11");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(5);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);

  await expect(reviewCard.locator(".card-text")).toHaveText(
    `Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always e...`
  );
  await reviewCard.getByTestId("full-review").click();

  await expect(page.locator("h1")).toHaveText("Weird: The Al Yankovic Story");
  await expect(page.locator(".poster")).toHaveAttribute(
    "src",
    "https://image.tmdb.org/t/p/w600_and_h900_bestv2/qcj2z13G0KjaIgc01ifiUKu7W07.jpg"
  );
  await expect(page.locator(".release-date")).toHaveText("Released: 9/8/2022");
  await expect(page.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );

  await expect(
    page.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(5);
  await expect(
    page.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);

  await expect(page.getByTestId("blurb")).toHaveText(
    `Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.

Daniel Radcliffe is fantastic, and it's a great film role for Rainn Wilson. There are a million great cameos.

You'll like it if you enjoy things like Children's Hospital, Comedy Bang Bang, or Popstar.`
  );
});

test("adds a new rating and fills all fields", async ({ page }) => {
  await page.getByTestId("add-rating").click();

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

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText(
    "Eternal Sunshine of the Spotless Mind"
  );
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2022-10-29");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(5);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
});

test("HTML tags in reviews are encoded properly", async ({ page }) => {
  await page.getByTestId("add-rating").click();

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

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("This is the <b>best</b> movie ever!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");
  await expect(reviewCard.locator(".card-title")).toHaveText(
    "Eternal Sunshine of the Spotless Mind"
  );
  await expect(
    (await reviewCard.locator(".card-text").innerHTML()).trim()
  ).toEqual("This is the &lt;b&gt;best&lt;/b&gt; movie ever!<br>");
});

test("adds a new rating and edits the details", async ({ page }) => {
  await page.getByTestId("add-rating").click();

  await page.locator("title-search #media-title").fill("something about mary");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("There's Something About Mary (1998)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "There's Something About Mary"
  );

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.getByTestId("edit-rating").click();

  await expect(page.locator("h1")).toHaveText(
    "There's Something About Mary (1998)"
  );

  await expect(page.locator("#rating-select")).toHaveValue("5");
  await page.locator("#rating-select").selectOption({ label: "4" });

  await expect(page.locator("#watched-date")).toHaveValue("2022-10-29");
  await page.locator("#watched-date").fill("2022-10-22");

  await expect(page.locator("#blurb")).toHaveValue("My favorite movie!");
  await page.locator("#blurb").fill("Not as good as I remembered");

  await page.locator("form .btn-primary").click();

  await expect(page).toHaveURL("/reviews");

  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2022-10-22");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(4);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(1);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Not as good as I remembered"
  );
});

test("adds a new rating and cancels the edit", async ({ page }) => {
  await page.getByTestId("add-rating").click();

  await page.locator("title-search #media-title").fill("the english pati");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("The English Patient (1996)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "The English Patient"
  );

  await page.locator("#rating-select").selectOption({ label: "4" });
  await page.locator("#watched-date").fill("2022-10-29");
  await page.locator("#blurb").fill("What an English patient he was!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(":nth-match(.card, 1)");

  await reviewCard.getByTestId("edit-rating").click();

  // Make edits that will be ignored when we cancel the edit.
  await page.locator("#rating-select").selectOption({ label: "1" });
  await page.locator("#watched-date").fill("2022-10-22");
  await page.locator("#blurb").fill("Ignore this edit");

  await page.getByTestId("cancel-edit").click();

  await expect(page).toHaveURL("/reviews");

  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2022-10-29");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(4);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(1);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "What an English patient he was!"
  );
});

test("editing another user's review fails", async ({ page, browser }) => {
  await page.getByTestId("add-rating").click();

  await page.locator("title-search #media-title").fill("the english pati");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("The English Patient (1996)");
  await matchingTitle.click();
  await expect(page.locator("title-search #media-title")).toHaveValue(
    "The English Patient"
  );

  await page.locator("#rating-select").selectOption({ label: "4" });
  await page.locator("#watched-date").fill("2022-10-29");
  await page.locator("#blurb").fill("What an English patient he was!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  // Switch to other user.
  const guestContext = await browser.newContext();

  // Share database across users.
  await guestContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const guestPage = await guestContext.newPage();
  await loginAsUserB(guestPage);

  const response = await guestPage.goto("/reviews/2/edit");
  await expect(response?.status()).toBe(403);

  await guestContext.close();
});

test("views a movie with an existing review and adds a new review", async ({
  page,
}) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  await expect(page).toHaveURL("/movies/1#review1");

  await page.getByTestId("add-rating").click();

  await expect(page).toHaveURL("/reviews/new?movieId=1");

  await expect(
    page.getByRole("heading", { name: "The Waterboy" })
  ).toBeVisible();

  await expect(page.locator("title-search")).not.toBeVisible();

  await page.locator("#rating-select").selectOption({ label: "5" });
  await page.locator("#watched-date").fill("2023-01-05");
  await page.locator("#blurb").fill("Relevant as ever");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");
});

test("adds a comment to an existing review", async ({ page }) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await page.locator("comment-form:first");
  await expect(page).toHaveURL("/movies/1#review1");
  await page.locator("#comment-btn").click();
  await page.keyboard.type("You sure do!");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL("/movies/1#comment2");

  const reviewDiv = await page.locator("#comment2");
  await expect(reviewDiv.getByRole("link")).toHaveText("userA");
  await expect(reviewDiv.getByTestId("relative-time")).toHaveText("just now");
  await expect(reviewDiv.locator("[data-sj-purpose='body']")).toHaveText(
    "You sure do!"
  );
});

test("removes leading and trailing whitespace from comments", async ({
  page,
}) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await page.locator("comment-form:first");
  await expect(page).toHaveURL("/movies/1#review1");
  await page.locator("#comment-btn").click();

  await page.keyboard.press("Enter");
  await page.keyboard.press("Space");
  await page.keyboard.type("  Yes, but can you strip my whitespace?   ");
  await page.keyboard.press("Enter");
  await page.keyboard.press("Space");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL("/movies/1#comment2");

  const reviewDiv = await page.locator("#comment2");
  await expect(reviewDiv.getByRole("link")).toHaveText("userA");
  await expect(reviewDiv.getByTestId("relative-time")).toHaveText("just now");
  await expect(reviewDiv.locator("[data-sj-purpose='body']")).toHaveText(
    "Yes, but can you strip my whitespace?",
    { useInnerText: true }
  );
});
