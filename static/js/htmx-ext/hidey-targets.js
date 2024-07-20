/*
Adapted from https://github.com/bigskysoftware/htmx-extensions/blob/492ee703bd43cf886e83de5ef88cc0a4bb8c5ef5/src/response-targets/response-targets.js
Used under BSD Zero Clause License
*/

/* global htmx */

(function () {
  /** @type {import("../htmx").HtmxInternalApi} */
  var api;

  /**
   * @param {HTMLElement} elt
   * @returns {HTMLElement | null}
   */
  function getErrorTarget(elt) {
    if (!elt) {
      return null;
    }

    var attrValue = api.getClosestAttributeValue(elt, "hx-target-error");
    if (!attrValue) {
      return null;
    }
    return api.querySelectorExt(elt, attrValue);
  }

  htmx.defineExtension("hidey-targets", {
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
        const srcEl = evt.target;

        const targetError = api.querySelectorExt(
          srcEl,
          srcEl.getAttribute("hx-target-error")
        );
        if (!targetError) {
          return;
        }
        const targetSuccess = api.querySelectorExt(
          srcEl,
          srcEl.getAttribute("hx-target")
        );
        if (!targetSuccess) {
          return true;
        }

        if (api.getClosestAttributeValue(srcEl, "hx-swap") === "textContent") {
          targetSuccess.innerHTML = "";
        }
        targetError.innerHTML = "";
      }
      if (
        name === "htmx:beforeSwap" &&
        evt.detail.xhr &&
        evt.detail.xhr.status !== 200
      ) {
        if (evt.detail.target) {
          if (evt.detail.xhr.getAllResponseHeaders().match(/HX-Retarget:/i)) {
            evt.detail.shouldSwap = true;
            return true;
          }
        }
        if (!evt.detail.requestConfig) {
          return true;
        }
        var target = getErrorTarget(evt.detail.requestConfig.elt);
        if (target) {
          evt.detail.shouldSwap = true;
          evt.detail.target = target;
        }
        return true;
      }
    },
  });
})();
