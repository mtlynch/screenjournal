import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";
import { loginAsUserA, loginAsUserB } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("index page renders card for movie review with comments", async ({
  page,
}) => {
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "The Waterboy" }),
  });
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userB watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2020-10-05");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(5);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText("I love water!");
  await expect(reviewCard.getByTestId("comment-count")).toHaveText(" 1");
});

test("index page renders card for movie review without comments", async ({
  page,
}) => {
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "Billy Madison" }),
  });
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userB watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2023-02-05");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(3);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "A staggering lack of water."
  );
  await expect(reviewCard.getByTestId("comment-count")).toHaveCount(0);
});

test("index page renders card for TV show review", async ({ page }) => {
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "Seinfeld (Season 1)" }),
  });
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2024-11-04");
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "I see what the fuss is about!"
  );
});

test("index page sorts cards based on desired sorting", async ({ page }) => {
  // By default, sort by watch date.
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "Seinfeld (Season 2)"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
    "Seinfeld (Season 1)"
  );
  await expect(page.locator(":nth-match(.card, 3) .card-title")).toHaveText(
    "Billy Madison"
  );
  await expect(page.locator(":nth-match(.card, 4) .card-title")).toHaveText(
    "The Waterboy"
  );

  // The sort dropdown shouldn't be visible yet.
  await expect(page.locator("#sort-by")).not.toBeVisible();

  // Clicking the sort button makes the sort dropdown visible.
  await page.locator("#sort-btn").click();
  await expect(page.locator("#sort-by")).toBeVisible();

  // Choose to sort by rating.
  await page.locator("#sort-by").selectOption("rating");

  // Verify the sorting is now by rating.
  await expect(page).toHaveURL("/reviews?sortBy=rating");
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "The Waterboy"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
    "Seinfeld (Season 1)"
  );
  await expect(page.locator(":nth-match(.card, 3) .card-title")).toHaveText(
    "Seinfeld (Season 2)"
  );
  await expect(page.locator(":nth-match(.card, 4) .card-title")).toHaveText(
    "Billy Madison"
  );

  // Choose to sort by watch date.
  await page.locator("#sort-btn").click();
  await page.locator("#sort-by").selectOption("watch-date");

  // Verify the sorting is now by watch date.
  await expect(page).toHaveURL("/reviews?sortBy=watch-date");
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "Seinfeld (Season 2)"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
    "Seinfeld (Season 1)"
  );
  await expect(page.locator(":nth-match(.card, 3) .card-title")).toHaveText(
    "Billy Madison"
  );
  await expect(page.locator(":nth-match(.card, 4) .card-title")).toHaveText(
    "The Waterboy"
  );
});

test("adds a new movie rating and fills in only required fields", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("slow lear");
  await page.getByText("Slow Learners (2015)").click();

  await expect(page).toHaveURL(
    "/reviews/new/write?tmdbId=333287&mediaType=movie"
  );
  await expect(
    page.getByRole("heading", { name: "Slow Learners" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "1.0" });

  await page.getByLabel("When did you watch?").fill("2022-10-21");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "Slow Learners" }),
  });
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
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(0);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(4);
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
});

test("adds a new TV show rating and fills in only required fields", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByLabel("TV show").click();

  await page.getByPlaceholder("Search").pressSequentially("30 r");
  await page.getByText("30 Rock (2006)").click();

  await expect(page).toHaveURL("/reviews/new/tv/pick-season?tmdbId=4608");
  await page.getByLabel("Season").selectOption({ label: "5" });

  await expect(page).toHaveURL(
    "/reviews/new/write?season=5&mediaType=tv-show&tmdbId=4608"
  );
  await expect(page.getByRole("heading", { name: "30 Rock" })).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "4.5" });

  await page.getByLabel("When did you watch?").fill("2024-10-05");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/tv-shows/2?season=5");

  await expect(page.getByRole("heading", { name: "30 Rock" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Season 5" })).toBeVisible();
  await expect(page.getByText("IMDB")).toHaveAttribute(
    "href",
    "https://www.imdb.com/title/tt0496424/"
  );

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "30 Rock (Season 5)" }),
  });
  await expect(reviewCard.locator(".card-subtitle")).toHaveText(
    /userA watched this .+ ago/,
    { useInnerText: true }
  );
  await expect(
    reviewCard.locator(".card-subtitle [data-testid='watch-date']")
  ).toHaveAttribute("title", "2024-10-05");
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(4);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
});

test("views a TV show with an existing review and adds a new review", async ({
  page,
}) => {
  await page
    .getByRole("heading", { name: "Seinfeld (Season 2)" })
    .getByRole("link")
    .click();

  await expect(page).toHaveURL("/tv-shows/1?season=2#review4");

  await page.getByRole("button", { name: "Add Rating" }).click();

  await expect(page).toHaveURL(
    "/reviews/new/write?season=2&mediaType=tv-show&tmdbId=1400"
  );

  await expect(page.getByRole("heading", { name: "Seinfeld" })).toBeVisible();

  await expect(page.getByPlaceholder("Search")).not.toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "5.0" });
  await page.getByLabel("When did you watch?").fill("2023-01-05");
  await page.getByLabel("Other thoughts?").fill("I liked it, too!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/tv-shows/1?season=2");
});

test("adds a new movie rating that's too long to display in a card", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page
    .getByPlaceholder("Search")
    .pressSequentially("Weird: The Al Yankovic Story");
  await page.getByText("Weird: The Al Yankovic Story (2022)").click();

  await expect(page).toHaveURL(
    "/reviews/new/write?tmdbId=928344&mediaType=movie"
  );
  await expect(
    page.getByRole("heading", { name: "Weird: The Al Yankovic Story" })
  ).toBeVisible();

  await page.getByLabel("When did you watch?").fill("2022-11-11");

  await page.getByLabel("Rating").selectOption({ label: "5.0" });

  await page.getByLabel("Other thoughts?")
    .fill(`Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.

Daniel Radcliffe is **fantastic**, and it's a great film role for Rainn Wilson. There are a million great cameos.

You'll like it if you enjoy things like _Children's Hospital_, _Comedy Bang Bang_, or _Popstar_.`);

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "Weird: The Al Yankovic Story" }),
  });
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
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(0);
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
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(0);
  await expect(
    page.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);

  await expect(
    (await page.getByTestId("blurb").innerHTML()).replace(/^\s+/gm, "")
  ).toEqual(
    `<p>Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.</p>
<p>If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.</p>
<p>Daniel Radcliffe is <strong>fantastic</strong>, and it's a great film role for Rainn Wilson. There are a million great cameos.</p>
<p>You'll like it if you enjoy things like <em>Children's Hospital</em>, <em>Comedy Bang Bang</em>, or <em>Popstar</em>.</p>
`
  );
});

test("adds a new rating and fills all fields", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("eternal sunshine");
  await page.getByText("Eternal Sunshine of the Spotless Mind (2004)").click();

  await expect(page).toHaveURL("/reviews/new/write?tmdbId=38&mediaType=movie");
  await expect(
    page.getByRole("heading", { name: "Eternal Sunshine of the Spotless Mind" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "5.0" });

  await page.getByLabel("When did you watch?").fill("2022-10-29");

  await page.getByLabel("Other thoughts?").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", {
      name: "Eternal Sunshine of the Spotless Mind",
    }),
  });
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
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(0);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
});

test("HTML tags in reviews are stripped from review excerpt", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("eternal sunshine");
  await page.getByText("Eternal Sunshine of the Spotless Mind (2004)").click();

  await expect(page).toHaveURL("/reviews/new/write?tmdbId=38&mediaType=movie");
  await expect(
    page.getByRole("heading", { name: "Eternal Sunshine of the Spotless Mind" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "5.0" });

  await page.getByLabel("When did you watch?").fill("2022-10-29");

  await page
    .getByLabel("Other thoughts?")
    .fill("This is the <b>best</b> movie ever!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", {
      name: "Eternal Sunshine of the Spotless Mind",
    }),
  });
  await expect(
    (await reviewCard.locator(".card-text").innerHTML()).trim()
  ).toEqual("This is the best movie ever!<br>");
});

test("adds a new movie rating and edits the details", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page
    .getByPlaceholder("Search")
    .pressSequentially("something about mary");
  await page.getByText("There's Something About Mary (1998)").click();

  await expect(page).toHaveURL("/reviews/new/write?tmdbId=544&mediaType=movie");
  await expect(
    page.getByRole("heading", { name: "There's Something About Mary" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "4.5" });

  await page.getByLabel("When did you watch?").fill("2022-10-29");

  await page.getByLabel("Other thoughts?").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "There's something About Mary" }),
  });
  await reviewCard.getByTestId("edit-rating").click();

  await expect(page).toHaveURL("/reviews/5/edit");

  await expect(
    page.getByRole("heading", { name: "There's Something About Mary (1998)" })
  ).toBeVisible();
  await expect(
    page.getByLabel("Rating").locator("option[selected]")
  ).toHaveText("4.5");

  await page.getByLabel("Rating").selectOption({ label: "3.5" });

  await expect(page.getByLabel("When did you watch?")).toHaveValue(
    "2022-10-29"
  );
  await page.getByLabel("When did you watch?").fill("2022-10-22");

  await expect(page.getByLabel("Other thoughts?")).toHaveValue(
    "My favorite movie!"
  );
  await page.getByLabel("Other thoughts?").fill("Not as good as I remembered");

  // No submit button on this form.
  await page.locator("form .btn-primary").click();

  await expect(page).toHaveURL("/movies/3#review5");

  await page.getByRole("menuitem", { name: "Home" }).click();
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
  ).toHaveCount(3);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(1);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "Not as good as I remembered"
  );
});

test("adds a new rating and cancels the edit", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("the english pati");
  await page.getByText("The English Patient (1996)").click();

  await expect(page).toHaveURL("/reviews/new/write?tmdbId=409&mediaType=movie");
  await expect(
    page.getByRole("heading", { name: "The English Patient" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "4.0" });
  await page.getByLabel("When did you watch?").fill("2022-10-29");
  await page
    .getByLabel("Other thoughts?")
    .fill("What an English patient he was!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");
  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "The English Patient" }),
  });
  await reviewCard.getByTestId("edit-rating").click();

  // Make edits that will be ignored when we cancel the edit.
  await page.getByLabel("Rating").selectOption({ label: "1.0" });
  await page.getByLabel("When did you watch?").fill("2022-10-22");
  await page.getByLabel("Other thoughts?").fill("Ignore this edit");

  await page.getByRole("button", { name: "Cancel" }).click();

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
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(0);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(1);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "What an English patient he was!"
  );
});

test("editing another user's review fails", async ({ page, browser }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("the english pati");
  await page.getByText("The English Patient (1996)").click();

  await expect(page).toHaveURL("/reviews/new/write?tmdbId=409&mediaType=movie");
  await expect(
    page.getByRole("heading", { name: "The English Patient" })
  ).toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "4.0" });
  await page.getByLabel("When did you watch?").fill("2022-10-29");
  await page
    .getByLabel("Other thoughts?")
    .fill("What an English patient he was!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/3");

  await page.getByRole("menuitem", { name: "Home" }).click();
  await expect(page).toHaveURL("/reviews");

  // Switch to other user.
  const guestContext = await browser.newContext();

  // Share database across users.
  await guestContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const guestPage = await guestContext.newPage();
  await guestPage.goto("/");
  await loginAsUserB(guestPage);

  const response = await guestPage.goto("/reviews/5/edit");
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

  await page.getByRole("button", { name: "Add Rating" }).click();

  await expect(page).toHaveURL("/reviews/new/write?movieId=1");

  await expect(
    page.getByRole("heading", { name: "The Waterboy" })
  ).toBeVisible();

  await expect(page.getByPlaceholder("Search")).not.toBeVisible();

  await page.getByLabel("Rating").selectOption({ label: "5.0" });
  await page.getByLabel("When did you watch?").fill("2023-01-05");
  await page.getByLabel("Other thoughts?").fill("Relevant as ever");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/movies/1");
});

test("adds a comment to an existing review", async ({ page }) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await expect(page).toHaveURL("/movies/1#review1");

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.getByRole("button", { name: "Comment" }).click();
  await expect(reviewDiv.getByRole("textbox")).toBeFocused();
  await page.keyboard.type("I loved it despite my indifference to water.");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");

  const commentDiv = await page.locator("#comment2");
  await expect(commentDiv.getByRole("link", { name: "userA" })).toBeVisible();
  await expect(commentDiv.getByTestId("relative-time")).toHaveText("just now");
  await expect(
    commentDiv.getByText("I loved it despite my indifference to water.")
  ).toBeVisible();
});

test("cancels a comment to an existing review", async ({ page }) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await expect(page).toHaveURL("/movies/1#review1");

  // Start a comment.
  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.getByRole("button", { name: "Comment" }).click();
  await expect(reviewDiv.getByRole("textbox")).toBeFocused();
  await page.keyboard.type("Lemme think about this...");

  // But then cancel.
  await reviewDiv.getByRole("button", { name: "Cancel" }).click();

  await expect(page).toHaveURL("/movies/1#review1");
  await expect(
    reviewDiv.getByRole("button", { name: "Comment" })
  ).toBeVisible();
});

test("adds, edits, and deletes a comment on an existing review", async ({
  page,
}) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  // Use a separate scope for variables on the current page.
  {
    const reviewDiv = await page.locator(".review", {
      has: page.getByText("I love water!"),
    });
    await reviewDiv.getByRole("button", { name: "Comment" }).click();
    await expect(reviewDiv.getByRole("textbox")).toBeFocused();

    await page.keyboard.type("We must ask ourselves...");
    await page.keyboard.press("Enter");
    await page.keyboard.press("Enter");
    await page.keyboard.type(`What is "movie?"`);
    await page.keyboard.press("Tab");
    await page.keyboard.press("Enter");

    const commentDiv = await page.locator("#comment2");
    await expect(commentDiv.getByRole("link", { name: "userA" })).toBeVisible();
    await expect(commentDiv.getByTestId("relative-time")).toHaveText(
      "just now"
    );
    await expect(
      // Strip leading whitespace from each line in inner HTML.
      (
        await commentDiv.locator("[data-sj-purpose='body']").innerHTML()
      ).replace(/^\s+/gm, "")
    ).toEqual(
      `<p>We must ask ourselves...</p>
<p>What is "movie?"</p>
`
    );

    await commentDiv.getByRole("link", { name: "Edit" }).click();
    await expect(reviewDiv.getByRole("textbox")).toBeFocused();

    // Select all in text field, and delete it.
    await page.keyboard.press("Control+A");
    await page.keyboard.press("Backspace");

    await page.keyboard.type("Actually, I thought this was meh.");
    await reviewDiv.getByRole("button", { name: "Save" }).click();
  }

  // Use a separate scope for variables on the current page.
  {
    const commentDiv = await page.locator("#comment2");
    await expect(commentDiv.locator("[data-sj-purpose='body']")).toHaveText(
      "Actually, I thought this was meh."
    );

    // Start to delete but then cancel.
    const dismissDialog = (dialog) => dialog.dismiss();
    page.on("dialog", dismissDialog);
    await commentDiv.getByRole("link", { name: "Delete" }).click();

    // Delete for real.
    page.removeListener("dialog", dismissDialog);
    page.on("dialog", (dialog) => dialog.accept());
    await commentDiv.getByRole("link", { name: "Delete" }).click();
  }

  // Verify comment2 is deleted.
  await expect(page).toHaveURL("/movies/1#review1");
  await expect(page.locator("#comment2")).toHaveCount(0);
});

test("allows edits to TV show review", async ({ page }) => {
  await page
    .locator(".card", {
      has: page.getByRole("heading", { name: "Seinfeld (Season 1)" }),
    })
    .getByText("Edit")
    .click();

  await expect(page).toHaveURL("/reviews/3/edit");
  await page.getByLabel("Rating").selectOption({ label: "4.5" });
  await page.getByLabel("When did you watch?").fill("2024-11-05");
  await page
    .getByLabel("Other thoughts?")
    .fill("On second thought: slightly overrated.");
  await page.getByRole("button", { name: "Save" }).click();

  await expect(page).toHaveURL("/tv-shows/1?season=1#review3");
  const reviewCard = await page.locator("div", {
    has: page.getByText("On second thought: slightly overrated."),
  });

  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-solid")
  ).toHaveCount(4);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star-half-stroke.fa-solid")
  ).toHaveCount(1);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
});

test("removes leading and trailing whitespace from comments", async ({
  page,
}) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await expect(page).toHaveURL("/movies/1#review1");

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.getByRole("button", { name: "Comment" }).click();
  await expect(reviewDiv.getByRole("textbox")).toBeFocused();

  await page.keyboard.press("Enter");
  await page.keyboard.type("Yes, but can you strip my whitespace?");
  await page.keyboard.press("Enter");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");

  const commentDiv = await page.locator("#comment2");
  await expect(commentDiv.getByRole("link", { name: "userA" })).toBeVisible();
  await expect(commentDiv.getByTestId("relative-time")).toHaveText("just now");
  await expect(
    // Strip leading whitespace from each line in inner HTML.
    (
      await commentDiv.locator("[data-sj-purpose='body']").innerHTML()
    ).replace(/^\s+/gm, "")
  ).toEqual(`<p>Yes, but can you strip my whitespace?</p>
`);
});

test("views reviews filtered by user", async ({ page }) => {
  await page
    .locator(".card", {
      has: page.getByRole("heading", { name: "The Waterboy" }),
    })
    .getByTestId("reviews-by-user")
    .click();

  await expect(page).toHaveURL("/reviews/by/userB");

  await expect(
    page.getByRole("heading", { name: "userB's ratings" })
  ).toBeVisible();

  await expect(page.getByText("userB has written 3 reviews")).toBeVisible();
});
