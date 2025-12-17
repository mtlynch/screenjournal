import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";
import { loginAsUserA, loginAsUserB, loginAsAdmin } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
});

test("adds an emoji reaction to a review", async ({ page }) => {
  await loginAsUserA(page);

  // Navigate to a movie page with an existing review.
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  // Find the review section.
  const reviewDiv = page.locator(".review").first();

  // Click the thumbs up emoji button.
  await reviewDiv.locator(".emoji-picker button", { hasText: "👍" }).click();

  // Verify reaction appears.
  const reactionDiv = reviewDiv.locator(".reaction", {
    hasText: /👍\s+userA/,
  });
  await expect(reactionDiv).toBeVisible();
  await expect(reactionDiv.getByTestId("relative-time")).toHaveText("just now");
});

test("emoji picker is hidden after user reacts", async ({ page }) => {
  await loginAsUserA(page);

  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = page.locator(".review").first();

  // Initially emoji picker is visible.
  await expect(reviewDiv.locator(".emoji-picker")).toBeVisible();

  // Add a reaction.
  await reviewDiv.locator(".emoji-picker button", { hasText: "🥞" }).click();

  // Emoji picker should be hidden.
  await expect(reviewDiv.locator(".emoji-picker")).not.toBeVisible();
});

test("user can delete their own reaction", async ({ page }) => {
  await loginAsUserA(page);

  // First add a reaction.
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = page.locator(".review").first();

  await reviewDiv.locator(".emoji-picker button", { hasText: "👀" }).click();

  // Verify reaction exists.
  const reactionDiv = reviewDiv.locator(".reaction", {
    hasText: /👀\s+userA/,
  });
  await expect(reactionDiv).toBeVisible();

  // Delete the reaction (accept confirmation dialog).
  page.on("dialog", (dialog) => dialog.accept());
  await reactionDiv.locator("button[data-sj-purpose='delete']").click();

  // Verify reaction is removed.
  await expect(reactionDiv).not.toBeVisible();

  // Reload to get the emoji menu back.
  await page.reload();

  // Verify emoji picker reappears.
  await expect(reviewDiv.locator(".emoji-picker")).toBeVisible();
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

  const reviewDiv = page.locator(".review").first();
  await reviewDiv.locator(".emoji-picker button", { hasText: "🤔" }).click();

  // Verify reaction was added.
  const reactionDiv = reviewDiv.locator(".reaction", {
    hasText: /🤔\s+userA/,
  });
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
  const userBReviewDiv = userBPage.locator(".review").first();
  const userBReactionDiv = userBReviewDiv.locator(".reaction", {
    hasText: /🤔\s+userA/,
  });
  await expect(userBReactionDiv).toBeVisible();
  await expect(
    userBReactionDiv.locator("button[data-sj-purpose='delete']")
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

  const reviewDiv = page.locator(".review").first();
  await reviewDiv.locator(".emoji-picker button", { hasText: "😯" }).click();

  // Verify reaction was added.
  const reactionDiv = reviewDiv.locator(".reaction", {
    hasText: /😯\s+userA/,
  });
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
  const adminReviewDiv = adminPage.locator(".review").first();
  const adminReactionDiv = adminReviewDiv.locator(".reaction", {
    hasText: /😯\s+userA/,
  });
  await expect(
    adminReactionDiv.locator("button[data-sj-purpose='delete']")
  ).toBeVisible();

  // Admin deletes userA's reaction.
  adminPage.on("dialog", (dialog) => dialog.accept());
  await adminReactionDiv.locator("button[data-sj-purpose='delete']").click();

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

  const reviewDiv = page.locator(".review").first();
  await reviewDiv.locator(".emoji-picker button", { hasText: "👍" }).click();

  // Verify first reaction was added.
  await expect(
    reviewDiv.locator(".reaction", { hasText: /👍\s+userA/ })
  ).toBeVisible();

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

  const userBReviewDiv = userBPage.locator(".review").first();
  await userBReviewDiv
    .locator(".emoji-picker button", { hasText: "🥞" })
    .click();

  // Verify reactions appear in chronological order (userA's first, then userB's).
  const reactions = userBReviewDiv.locator(".reaction");
  await expect(reactions).toHaveCount(2);
  await expect(reactions.first()).toContainText("userA");
  await expect(reactions.last()).toContainText("userB");

  await userBContext.close();
});
