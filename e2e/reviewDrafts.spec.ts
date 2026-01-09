import { test, expect } from "@playwright/test";
import { populateDummyData } from "./helpers/db";
import { loginAsUserA } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("auto-saves a draft and lets the user resume", async ({ page }) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("slow lear");
  await page.getByText("Slow Learners (2015)").click();

  await page.getByLabel("When did you watch?").fill("2024-10-10");

  const draftResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/reviews/drafts") &&
      response.request().method() === "POST" &&
      response.status() === 201
  );
  await page.getByLabel("Other thoughts?").fill("Draft thoughts");
  await draftResponse;

  await page.getByRole("menuitem", { name: "Account" }).click();
  await page.getByRole("menuitem", { name: "Drafts" }).click();

  await expect(page).toHaveURL("/reviews/drafts");

  const draftCard = page.getByTestId("draft-card").first();
  await expect(draftCard.getByText("Slow Learners")).toBeVisible();
  await draftCard.getByRole("link", { name: "Continue" }).click();

  await expect(page.getByLabel("Other thoughts?")).toHaveValue(
    "Draft thoughts"
  );

  await page.getByRole("button", { name: "Publish" }).click();
  await expect(page).toHaveURL("/movies/3");
});

test("redirects to an existing draft when starting the same review", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("slow lear");
  await page.getByText("Slow Learners (2015)").click();

  await page.getByLabel("When did you watch?").fill("2024-10-10");

  const draftResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/reviews/drafts") &&
      response.request().method() === "POST" &&
      response.status() === 201
  );
  await page.getByLabel("Other thoughts?").fill("Second draft");
  await draftResponse;

  await page.getByRole("menuitem", { name: "Home" }).click();
  await page.getByRole("button", { name: "Add Rating" }).click();

  await page.getByPlaceholder("Search").pressSequentially("slow lear");
  await page.getByText("Slow Learners (2015)").click();

  await expect(page).toHaveURL(/\/reviews\/\d+\/edit/);
  await expect(page.getByText("Draft")).toBeVisible();
});
