import { test, expect } from "@playwright/test";
import { populateDummyData, wipeDB } from "./helpers/db.js";
import { loginAsUserA } from "./helpers/login.js";

test.beforeEach(async ({ page }) => {
  await wipeDB(page);
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("TODO", async ({ page }) => {
  //TODO
});
