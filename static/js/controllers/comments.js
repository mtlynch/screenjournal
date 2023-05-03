import { processJsonResponse, processPlaintextResponse } from "./common.js";

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

export function commentPut(commentId, comment) {
  return fetch(`/api/comments/${commentId}`, {
    method: "PUT",
    credentials: "include",
    body: JSON.stringify({
      commentId,
      comment: comment.trim(),
    }),
  }).then(processPlaintextResponse);
}

export function commentDelete(commentId) {
  return fetch(`/api/comments/${commentId}`, {
    method: "DELETE",
    credentials: "include",
  }).then(processPlaintextResponse);
}
