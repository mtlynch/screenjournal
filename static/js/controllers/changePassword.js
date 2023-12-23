export async function changePassword(oldPassword, newPassword) {
  return fetch("/api/auth/change-password", {
    method: "POST",
    mode: "same-origin",
    credentials: "include",
    cache: "no-cache",
    redirect: "error",
    body: JSON.stringify({
      oldPassword,
      newPassword,
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
