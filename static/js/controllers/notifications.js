import { processPlaintextResponse } from "./common.js";

export function notificationsPost(isSubscribedToNewReviews) {
  return fetch(`/api/account/notifications`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify({ isSubscribedToNewReviews }),
  }).then(processPlaintextResponse);
}
