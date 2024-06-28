import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";
import { loginAsUserA, loginAsUserB } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("index page renders card for review with comments", async ({ page }) => {
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

test("index page renders card for review without comments", async ({
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
  ).toHaveCount(2);
  await expect(
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(3);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "A staggering lack of water."
  );
  await expect(reviewCard.getByTestId("comment-count")).toHaveCount(0);
});

test("index page sorts cards based on desired sorting", async ({ page }) => {
  // By default, sort by watch date.
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "Billy Madison"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
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
  await expect(page).toHaveURL("/reviews?sort-by=rating");
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "The Waterboy"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
    "Billy Madison"
  );

  // Choose to sort by rating.
  await page.locator("#sort-btn").click();
  await page.locator("#sort-by").selectOption("watch-date");

  // Verify the sorting is now by watch date.
  await expect(page).toHaveURL("/reviews?sort-by=watch-date");
  await expect(page.locator(":nth-match(.card, 1) .card-title")).toHaveText(
    "Billy Madison"
  );
  await expect(page.locator(":nth-match(.card, 2) .card-title")).toHaveText(
    "The Waterboy"
  );
});

test("adds a new rating and fills in only required fields", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("slow lear");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("Slow Learners (2015)");
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue("Slow Learners");

  await page.locator("#rating-select").selectOption({ label: "1" });

  await page.locator("#watched-date").fill("2022-10-21");

  await page.locator("form input[type='submit']").click();

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
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(4);
  await expect(reviewCard.locator(".card-text")).toHaveCount(0);
});

test("adds a new rating that's too long to display in a card", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("Weird: The Al Yankovic Story");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("Weird: The Al Yankovic Story (2022)");
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
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
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("eternal sunshine");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText(
    "Eternal Sunshine of the Spotless Mind (2004)"
  );
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
    "Eternal Sunshine of the Spotless Mind"
  );

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

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
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(0);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "My favorite movie!"
  );
});

test("HTML tags in reviews are encoded properly", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("eternal sunshine");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText(
    "Eternal Sunshine of the Spotless Mind (2004)"
  );
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
    "Eternal Sunshine of the Spotless Mind"
  );

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("This is the <b>best</b> movie ever!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", {
      name: "Eternal Sunshine of the Spotless Mind",
    }),
  });
  await expect(
    (await reviewCard.locator(".card-text").innerHTML()).trim()
  ).toEqual("This is the &lt;b&gt;best&lt;/b&gt; movie ever!<br>");
});

test("adds a new rating and edits the details", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("something about mary");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("There's Something About Mary (1998)");
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
    "There's Something About Mary"
  );

  await page.locator("#rating-select").selectOption({ label: "5" });

  await page.locator("#watched-date").fill("2022-10-29");

  await page.locator("#blurb").fill("My favorite movie!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "There's something About Mary" }),
  });
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
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("the english pati");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("The English Patient (1996)");
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
    "The English Patient"
  );

  await page.locator("#rating-select").selectOption({ label: "4" });
  await page.locator("#watched-date").fill("2022-10-29");
  await page.locator("#blurb").fill("What an English patient he was!");

  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/reviews");

  const reviewCard = await page.locator(".card", {
    has: page.getByRole("heading", { name: "The English Patient" }),
  });
  await reviewCard.getByTestId("edit-rating").click();

  // Make edits that will be ignored when we cancel the edit.
  await page.locator("#rating-select").selectOption({ label: "1" });
  await page.locator("#watched-date").fill("2022-10-22");
  await page.locator("#blurb").fill("Ignore this edit");

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
    reviewCard.locator("[data-testid='rating'] .fa-star.fa-regular")
  ).toHaveCount(1);
  await expect(reviewCard.locator(".card-text")).toHaveText(
    "What an English patient he was!"
  );
});

test("editing another user's review fails", async ({ page, browser }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").fill("the english pati");
  const matchingTitle = await page.locator(
    "#search-results-list li:first-child span"
  );
  await expect(matchingTitle).toHaveText("The English Patient (1996)");
  await matchingTitle.click();
  await expect(page.getByPlaceholder("Search")).toHaveValue(
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

  const response = await guestPage.goto("/reviews/3/edit");
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

  await expect(page).toHaveURL("/reviews/new?movieId=1");

  await expect(
    page.getByRole("heading", { name: "The Waterboy" })
  ).toBeVisible();

  await expect(page.getByPlaceholder("Search")).not.toBeVisible();

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
      `
We must ask ourselves...<br>
<br>
What is "movie?"<br>
`.trimStart()
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
  ).toEqual(
    `
    Yes, but can you strip my whitespace?<br>
`.trimStart()
  );
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

  await expect(page.getByTestId("collection-count")).toHaveText(
    "userB has reviewed 2 movies"
  );
});
