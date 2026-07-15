/* global htmx */

// Keep htmx requests limited to this origin so hx-* attributes can't be used to
// send data to another site.
htmx.config.selfRequestsOnly = true;

// Server-rendered htmx responses don't need response-provided script execution.
htmx.config.allowScriptTags = false;

// Keep htmx from evaluating dynamic JavaScript strings.
htmx.config.allowEval = false;

// Disable htmx's history cache to avoid stale pages and storing page contents in
// long-lived browser storage.
htmx.config.historyCacheSize = 0;

// History restores should fetch full pages, not HX-Request partial responses.
htmx.config.historyRestoreAsHxRequest = false;

// Indicator CSS lives in screenjournal.css, so htmx shouldn't inject its own
// inline style tag.
htmx.config.includeIndicatorStyles = false;

// Fail stalled requests instead of leaving controls disabled indefinitely.
htmx.config.timeout = 5000;

// Don't let response-targets override isError.
htmx.config.responseTargetUnsetsError = false;

// Module scripts run after the document is parsed, so document.body exists.
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
