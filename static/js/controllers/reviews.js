import { processPlaintextResponse } from "./common.js";

export function deleteReview(id) {
  return fetch(`/api/reviews/${id}`, {
    method: "DELETE",
    credentials: "include",
    headers: {
      Accept: "text/plain",
    },
  }).then(processPlaintextResponse);
}
