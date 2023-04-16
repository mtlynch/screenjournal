export async function populateDummyData(page) {
  await page.goto("/"); // hack to populate the DB cookie
  await page.goto("/api/debug/db/populate-dummy-data");
}
