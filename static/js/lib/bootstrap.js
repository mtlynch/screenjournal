export function makeInvisible(el) {
  el.classList.add("invisible");
}

export function makeVisible(el) {
  el.classList.remove("invisible");
}

export function hideElement(el) {
  el.classList.add("d-none");
}

export function showElement(el) {
  el.classList.remove("d-none");
}
