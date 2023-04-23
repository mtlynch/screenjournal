import { processJsonResponse } from "./common.js";

export function commentPost(reviewId, comment) {
  return fetch(`/api/reviews/${reviewId}/comment`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify({ comment: comment.trim() }),
  })
    .then(processJsonResponse)
    .then((result) => {
      return Promise.resolve(result.id);
    });
}
