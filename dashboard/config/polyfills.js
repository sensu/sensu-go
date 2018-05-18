/* eslint-disable import/no-dynamic-require */
/* eslint-disable global-require */

// fetch() polyfill for making API calls.
require("whatwg-fetch");

// URLSearch
require("url-search-params-polyfill");

// Intl.RelativeTimeFormat
window.IntlRelativeFormat = require("intl-relativeformat"); // eslint-disable-line no-unused-vars
require("intl-relativeformat/dist/locale-data/en.js");
