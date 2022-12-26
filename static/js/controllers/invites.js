import { processPlaintextResponse } from "./common.js";

export function createInvite(invitee) {
  return fetch(`/api/admin/invites`, {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
    body: JSON.stringify({
      invitee: invitee,
    }),
  }).then(processPlaintextResponse);
}
