import { test, expect } from "@playwright/test";
import { registerAsAdmin } from "./helpers/auth.js";
import { wipeDB } from "./helpers/wipe.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await registerAsAdmin(page);
});

test("signing up with an valid invite code succeeds", async ({
  page,
  browser,
}) => {
  await page.locator("id=admin-dropdown").click();
  await page.locator("data-test-id=invites-btn").click();

  await expect(page).toHaveURL("/admin/invites");
  await page.locator("data-test-id=create-invite").click();

  await expect(page).toHaveURL("/admin/invites/new");
  await expect(page.locator("id=invitee")).toBeFocused();
  await page.locator("id=invitee").fill("Billy");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/admin/invites");

  const inviteLink =
    (await page.locator("data-test-id=invite-link").getAttribute("href")) || "";

  const guestContext = await browser.newContext();
  const guestPage = await guestContext.newPage();
  await guestPage.goto(inviteLink);

  await expect(guestPage.locator(".alert-info")).toHaveText(
    "Welcome, Billy! We've been expecting you."
  );
  await expect(guestPage.locator("id=username")).toHaveValue("billy");
  await guestPage.locator("id=username").fill("billy123");
  await guestPage.locator("id=email").fill("billy@example.com");
  await guestPage.locator("id=password").fill("billypass");
  await guestPage.locator("id=password-confirm").fill("billypass");
  await guestPage.locator("form input[type='submit']").click();

  await expect(guestPage).toHaveURL("/reviews");
  await guestContext.close();
});

test("signing up with an invalid invite code fails", async ({
  page,
  browser,
}) => {
  await page.locator("id=admin-dropdown").click();
  await page.locator("data-test-id=invites-btn").click();

  await expect(page).toHaveURL("/admin/invites");
  await page.locator("data-test-id=create-invite").click();

  await expect(page).toHaveURL("/admin/invites/new");
  await expect(page.locator("id=invitee")).toBeFocused();
  await page.locator("id=invitee").fill("Nigel");
  await page.locator("form input[type='submit']").click();

  await expect(page).toHaveURL("/admin/invites");

  const guestContext = await browser.newContext();
  const guestPage = await guestContext.newPage();
  const response = await guestPage.goto("/sign-up?invite=1nv4LD");
  await expect(response?.status()).toBe(401);

  await guestContext.close();
});
