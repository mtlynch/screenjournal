htmx.config.selfRequestsOnly = true;
htmx.config.allowScriptTags = false;
htmx.config.allowEval = false;

document.body.addEventListener("htmx:beforeSwap", function (evt) {
  if (evt.detail.xhr.status === 204) {
    evt.detail.shouldSwap = true;
  }
});

// Inspired by https://stackoverflow.com/a/76134033/90388
htmx.defineExtension("reset-on-success", {
  onEvent: function (name, event) {
    if (name !== "htmx:beforeSwap" || event.detail.isError) {
      return;
    }

    const triggeringElt = event.detail.requestConfig.elt;
    if (
      !triggeringElt.closest("[hx-reset-on-success]") &&
      !triggeringElt.closest("[data-hx-reset-on-success]")
    ) {
      return;
    }

    switch (triggeringElt.tagName) {
      case "TEXTAREA":
        triggeringElt.value = triggeringElt.defaultValue;
        break;
      case "FORM":
        triggeringElt.reset();
        break;
    }
  },
});
