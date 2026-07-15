/* global htmx */

htmx.config.selfRequestsOnly = true; // Prevent hx-* from calling other origins.
htmx.config.allowScriptTags = false; // Server-rendered swaps don't need scripts.
htmx.config.allowEval = false; // Keep htmx from evaluating dynamic JS strings.
htmx.config.historyCacheSize = 0; // Avoid localStorage page-cache staleness/leaks.
htmx.config.historyRestoreAsHxRequest = false; // Restore history with full pages.
htmx.config.includeIndicatorStyles = false; // Indicator CSS lives in screenjournal.css.
htmx.config.timeout = 5000; // Fail stalled requests instead of disabling UI forever.
htmx.config.responseTargetUnsetsError = false; // Keep response-targets error state.
htmx.config.responseHandling = [
  { code: "204", swap: true }, // Empty 204 deletes clear the target element.
  { code: "422", swap: true, error: false }, // Validation errors swap normally.
  { code: "[23]..", swap: true }, // Successful non-empty responses swap normally.
  { code: "[45]..", swap: false, error: true }, // Error targets handle failures.
];
