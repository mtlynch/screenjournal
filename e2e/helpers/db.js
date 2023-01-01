export async function wipeDB(page) {
  await page.goto("/api/debug/db/wipe");
}

export async function populateDummyData(page) {
  await page.goto("/api/debug/db/populate-dummy-data");
}
