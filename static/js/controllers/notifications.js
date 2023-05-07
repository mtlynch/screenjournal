import { processPlaintextResponse } from "./common.js";

export function notificationsPost(
  isSubscribedToNewReviews,
  isSubscribedToAllComments
) {
  return fetch(`/api/account/notifications`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify({
      isSubscribedToNewReviews,
      isSubscribedToAllComments,
    }),
  }).then(processPlaintextResponse);
}
