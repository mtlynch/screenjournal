export async function wipeDB(page) {
  await page.goto("/api/debug/wipe-db");
}
