export function processJsonResponse(response) {
  if (response.ok) {
    return response.json();
  }
  return response.text().then((error) => {
    return Promise.reject(error);
  });
}
