export function makeInvisible(el) {
  el.classList.add("invisible");
}

export function makeVisible(el) {
  el.classList.remove("invisible");
}

export function reallyHideElement(el) {
  el.classList.add("d-none");
}
