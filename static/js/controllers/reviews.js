import { processPlaintextResponse } from "./common.js";

export function addReview(review) {
  return fetch(`/api/reviews`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify(review),
  }).then(processPlaintextResponse);
}
