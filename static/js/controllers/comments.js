import { processJsonResponse } from "./common.js";

export function commentPost(reviewId, comment) {
  return fetch("/api/comments", {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify({
      reviewId,
      comment: comment.trim(),
    }),
  })
    .then(processJsonResponse)
    .then((result) => {
      return Promise.resolve(result.id);
    });
}
