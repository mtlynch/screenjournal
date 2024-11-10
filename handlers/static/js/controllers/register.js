export async function register(email, username, password, inviteCode) {
  return fetch(`/api/users/${username}`, {
    method: "PUT",
    mode: "same-origin",
    credentials: "include",
    cache: "no-cache",
    redirect: "error",
    body: JSON.stringify({
      email: email,
      password: password,
      inviteCode: inviteCode,
    }),
  }).then((response) => {
    if (!response.ok) {
      return response.text().then((error) => {
        return Promise.reject(error);
      });
    }
    return Promise.resolve();
  });
}
