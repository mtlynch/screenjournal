import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";
import { loginAsUserA, loginAsUserB, loginAsAdmin } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
});

test("adds an emoji reaction to a review and deletes it", async ({ page }) => {
  await loginAsUserA(page);

  // Navigate to a movie page with an existing review.
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  // Find the review section by the reviewer's visible text.
  const reviewDiv = page.locator("div", {
    has: page.getByText("userB watched this", { exact: false }),
  });

  // Click the thumbs up emoji button.
  await reviewDiv.getByRole("button", { name: "👍" }).click();

  // Verify reaction appears.
  const reactionDiv = reviewDiv.getByText(/👍\s+userA/);
  await expect(reactionDiv).toBeVisible();
  await expect(
    reactionDiv.locator("..").getByTestId("relative-time")
  ).toHaveText("just now");

  // Emoji picker should be hidden (emoji buttons no longer visible).
  await expect(reviewDiv.getByRole("button", { name: "👍" })).not.toBeVisible();

  // Delete the reaction (accept confirmation dialog).
  page.on("dialog", (dialog) => dialog.accept());
  await reactionDiv.locator("..").getByTitle("Delete reaction").click();

  // Verify reaction is removed.
  await expect(reactionDiv).not.toBeVisible();

  // Reload to get the emoji menu back.
  await page.reload();

  // Verify emoji picker reappears (emoji buttons visible again).
  await expect(reviewDiv.getByRole("button", { name: "👍" })).toBeVisible();
});

test("user cannot delete another user's reaction", async ({
  page,
  browser,
}) => {
  // UserA adds a reaction.
  await loginAsUserA(page);
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = page.locator("div", {
    has: page.getByText("userB watched this", { exact: false }),
  });
  await reviewDiv.getByRole("button", { name: "🤔" }).click();

  // Verify reaction was added.
  const reactionDiv = reviewDiv.getByText(/🤔\s+userA/);
  await expect(reactionDiv).toBeVisible();

  // Switch to userB.
  const userBContext = await browser.newContext();
  await userBContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const userBPage = await userBContext.newPage();
  await userBPage.goto("/");
  await loginAsUserB(userBPage);

  await userBPage
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  // UserB should see userA's reaction but NOT see delete button.
  const userBReviewDiv = userBPage.locator("div", {
    has: userBPage.getByText("userB watched this", { exact: false }),
  });
  const userBReactionDiv = userBReviewDiv.getByText(/🤔\s+userA/);
  await expect(userBReactionDiv).toBeVisible();
  await expect(
    userBReactionDiv.locator("..").getByTitle("Delete reaction")
  ).not.toBeVisible();

  await userBContext.close();
});

test("admin can delete another user's reaction", async ({ page, browser }) => {
  // UserA adds a reaction.
  await loginAsUserA(page);
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = page.locator("div", {
    has: page.getByText("userB watched this", { exact: false }),
  });
  await reviewDiv.getByRole("button", { name: "😯" }).click();

  // Verify reaction was added.
  const reactionDiv = reviewDiv.getByText(/😯\s+userA/);
  await expect(reactionDiv).toBeVisible();

  // Switch to admin user.
  const adminContext = await browser.newContext();
  await adminContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const adminPage = await adminContext.newPage();
  adminPage.goto("/");
  await loginAsAdmin(adminPage);

  await adminPage
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  // Admin should see delete button on userA's reaction.
  const adminReviewDiv = adminPage.locator("div", {
    has: adminPage.getByText("userB watched this", { exact: false }),
  });
  const adminReactionDiv = adminReviewDiv.getByText(/😯\s+userA/);
  await expect(
    adminReactionDiv.locator("..").getByTitle("Delete reaction")
  ).toBeVisible();

  // Admin deletes userA's reaction.
  adminPage.on("dialog", (dialog) => dialog.accept());
  await adminReactionDiv.locator("..").getByTitle("Delete reaction").click();

  await expect(adminReactionDiv).not.toBeVisible();

  await adminContext.close();
});

test("reactions are displayed in chronological order", async ({
  page,
  browser,
}) => {
  // UserA adds first reaction.
  await loginAsUserA(page);
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = page.locator("div", {
    has: page.getByText("userB watched this", { exact: false }),
  });
  await reviewDiv.getByRole("button", { name: "👍" }).click();

  // Verify first reaction was added.
  await expect(reviewDiv.getByText(/👍\s+userA/)).toBeVisible();

  // UserB adds second reaction.
  const userBContext = await browser.newContext();
  await userBContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const userBPage = await userBContext.newPage();
  await userBPage.goto("/");
  await loginAsUserB(userBPage);
  await userBPage
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const userBReviewDiv = userBPage.locator("div", {
    has: userBPage.getByText("userB watched this", { exact: false }),
  });
  await userBReviewDiv.getByRole("button", { name: "🥞" }).click();

  // Verify reactions appear in chronological order (userA's first, then userB's).
  // Scope to reactions-section to exclude comments which also have relative-time.
  const reactions = userBReviewDiv
    .getByTestId("reactions-section")
    .getByTestId("relative-time");
  await expect(reactions).toHaveCount(2);
  await expect(reactions.nth(0).locator("..")).toContainText("userA");
  await expect(reactions.nth(1).locator("..")).toContainText("userB");

  await userBContext.close();
});
