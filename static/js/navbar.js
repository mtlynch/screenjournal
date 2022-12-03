import { logOut } from "./controllers/auth.js";

document.addEventListener("DOMContentLoaded", () => {
  const logOutEl = document.getElementById("navbar-log-out");
  if (logOutEl) {
    logOutEl.addEventListener("click", (evt) => {
      evt.preventDefault();

      logOut().then(() => {
        document.location = "/";
      });
    });
  }
});
