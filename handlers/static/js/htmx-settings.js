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

htmx.config.responseHandling = [
  // Empty 204 responses from delete endpoints should clear their target.
  { code: "204", swap: true },

  // Validation errors should swap normally without console error noise.
  { code: "422", swap: true, error: false },

  // Successful non-empty responses should swap normally.
  { code: "[23]..", swap: true },

  // Let response-targets route error responses to hx-target-error elements.
  { code: "[45]..", swap: false, error: true },
];
