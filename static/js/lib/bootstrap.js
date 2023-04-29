export function hideElement(el) {
  el.classList.add("invisible");
}

export function showElement(el) {
  el.classList.remove("invisible");
}

export function reallyHideElement(el) {
  el.classList.add("d-none");
}
