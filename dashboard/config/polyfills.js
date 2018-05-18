/* eslint-disable import/no-dynamic-require */
/* eslint-disable global-require */

if (typeof Promise === "undefined") {
  // Rejection tracking prevents a common issue where React gets into an
  // inconsistent state due to an error, but it gets swallowed by a Promise,
  // and the user has no idea what causes React's erratic future behavior.
  require("promise/lib/rejection-tracking").enable();
  window.Promise = require("promise/lib/es6-extensions.js");
}

// fetch() polyfill for making API calls.
require("whatwg-fetch");

// URLSearch
require("url-search-params-polyfill");

// Intl.RelativeTimeFormat
window.IntlRelativeFormat = require("intl-relativeformat"); // eslint-disable-line no-unused-vars
require("intl-relativeformat/dist/locale-data/en.js");
