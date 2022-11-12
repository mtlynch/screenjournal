import { processJsonResponse } from "./common.js";

export function searchGet(query) {
  return fetch(
    `/api/search?` +
      new URLSearchParams({
        query: query,
      }),
    {
      method: "GET",
      credentials: "include",
      headers: {
        Accept: "application/json",
      },
    }
  )
    .then(processJsonResponse)
    .then((result) => {
      return Promise.resolve(result);
    });
}
