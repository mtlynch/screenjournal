import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsAdmin } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsAdmin(page);
});

test("activity page shows reviews, comments, and reactions with links", async ({
  page,
}) => {
  await page.goto("/activity");

  await expect(page.getByRole("heading", { name: "Activity" })).toBeVisible();

  await expect(
    page.getByText("userB reviewed The Waterboy as")
  ).toBeVisible();
  await expect(
    page.getByText("userA reacted to userB's The Waterboy review with 🥞.")
  ).toBeVisible();
  await expect(
    page.getByText("userA replied to userB's review of The Waterboy.")
  ).toBeVisible();

  await expect(page.getByRole("link", { name: "userA" })).toHaveAttribute(
    "href",
    "/reviews/by/userA"
  );
  await expect(page.getByRole("link", { name: "userB" })).toHaveAttribute(
    "href",
    "/reviews/by/userB"
  );

  const waterboyReviewLink = page.getByRole("link", { name: "The Waterboy" });
  await expect(waterboyReviewLink).toHaveAttribute(
    "href",
    "/movies/1#review1"
  );

  const waterboyReviewItem = page
    .locator("li")
    .filter({ hasText: "userB reviewed The Waterboy as" });
  await expect(
    waterboyReviewItem.locator(".fa-solid.fa-star")
  ).toHaveCount(5);
});
