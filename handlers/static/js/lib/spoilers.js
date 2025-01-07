export function addSpoilersRevealButtons() {
  document.querySelectorAll(".spoilers").forEach((el) => {
    const btn = document.createElement("button");
    btn.classList.add("btn");
    btn.classList.add("btn-warning");
    btn.textContent = "Show Spoilers";
    el.after(btn);

    btn.addEventListener("click", () => {
      el.classList.remove("d-none");
      btn.remove();
    });
  });
}
