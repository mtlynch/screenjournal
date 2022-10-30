export function processPlaintextResponse(response) {
  if (response.ok) {
    return response.text();
  }
  return response.text().then((error) => {
    return Promise.reject(error);
  });
}
