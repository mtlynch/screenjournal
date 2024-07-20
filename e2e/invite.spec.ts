import { test, expect } from "@playwright/test";
import { loginAsAdmin } from "./helpers/login.js";
import { populateDummyData, readDbTokenCookie } from "./helpers/db.js";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsAdmin(page);
});

test("signing up with a valid invite code succeeds", async ({
  page,
  browser,
}) => {
  await page.getByRole("menuitem", { name: "Admin" }).click();
  await page.getByRole("menuitem", { name: "Invites" }).click();

  await expect(page).toHaveURL("/admin/invites");
  await expect(page.getByLabel("Invitee's name")).toBeFocused();
  await page.getByLabel("Invitee's name").fill("Billy");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/admin/invites");

  const inviteLink =
    (await page.getByTestId("invite-link").getAttribute("href")) || "";

  const guestContext = await browser.newContext();

  // Share database across users.
  await guestContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);

  const guestPage = await guestContext.newPage();
  await guestPage.goto(inviteLink);

  await expect(guestPage.locator(".alert-info")).toHaveText(
    "Welcome, Billy! We've been expecting you."
  );
  await expect(guestPage.getByLabel("Username")).toHaveValue("billy");
  await guestPage.getByLabel("Username").fill("billy123");
  await guestPage.getByLabel("Email Address").fill("billy@example.com");
  await guestPage.getByLabel("Password", { exact: true }).fill("billypass");
  await guestPage.getByLabel("Confirm Password").fill("billypass");
  await guestPage.locator("form input[type='submit']").click();

  await expect(guestPage).toHaveURL("/reviews");
  await guestContext.close();
});

test("signing up with an invalid invite code fails", async ({
  page,
  browser,
}) => {
  await page.getByRole("menuitem", { name: "Admin" }).click();
  await page.getByRole("menuitem", { name: "Invites" }).click();

  await expect(page).toHaveURL("/admin/invites");
  await expect(page.getByLabel("Invitee's name")).toBeFocused();
  await page.getByLabel("Invitee's name").fill("Nigel");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/admin/invites");

  const guestContext = await browser.newContext();

  // Share database across users.
  await guestContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);

  const guestPage = await guestContext.newPage();
  const response = await guestPage.goto("/sign-up?invite=222333");
  await expect(response?.status()).toBe(401);

  await guestContext.close();
});
