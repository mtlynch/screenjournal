/* global htmx */

// Tighten security.
htmx.config.selfRequestsOnly = true;
htmx.config.allowScriptTags = false;
htmx.config.allowEval = false;

// Don't let response-targets override isError.
htmx.config.responseTargetUnsetsError = false;

document.addEventListener("DOMContentLoaded", () => {
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (evt.detail.xhr.status === 204) {
      evt.detail.shouldSwap = true;
    }
    if (evt.detail.xhr.status === 422) {
      // allow 422 responses to swap as we are using this as a signal that
      // a form was submitted with bad data and want to rerender with the
      // errors
      //
      // set isError to false to avoid error logging in console
      evt.detail.shouldSwap = true;
      evt.detail.isError = false;
    }
  });
});
