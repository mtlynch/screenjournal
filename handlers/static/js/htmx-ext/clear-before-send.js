/* global htmx */

(function () {
  /** @type {import("../htmx").HtmxInternalApi} */
  let api;

  htmx.defineExtension("clear-before-send", {
    /** @param {import("../htmx").HtmxInternalApi} apiRef */
    init: function (apiRef) {
      api = apiRef;
    },

    /**
     * @param {string} name
     * @param {Event} evt
     */
    onEvent: function (name, evt) {
      if (name == "htmx:beforeSend") {
        const clearTargetRaw = evt.target.getAttribute("hx-clear");
        if (!clearTargetRaw) {
          return;
        }

        clearTargetRaw
          .split(",")
          .map((item) => item.trim())
          .forEach((clearTargetSelector) => {
            api.querySelectorExt(clearTargetSelector).innerHTML = "";
          });
      }
    },
  });
})();
