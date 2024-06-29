import { processPlaintextResponse } from "./common.js";

export function updateReview(id, review) {
  return fetch(`/api/reviews/${id}`, {
    method: "PUT",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify(review),
  }).then(processPlaintextResponse);
}

export function deleteReview(id) {
  return fetch(`/api/reviews/${id}`, {
    method: "DELETE",
    credentials: "include",
    headers: {
      Accept: "text/plain",
    },
  }).then(processPlaintextResponse);
}
