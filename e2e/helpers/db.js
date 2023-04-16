export async function populateDummyData(page) {
  await page.goto("/"); // hack to populate the DB cookie
  await page.goto("/api/debug/db/populate-dummy-data");
}

export function readDbTokenCookie(cookies) {
  for (const cookie of cookies) {
    if (cookie.name === "db-token") {
      return cookie;
    }
  }
  return undefined;
}
