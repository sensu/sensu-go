import modernizr from "/modernizr.macro";

// Configure bluebird polyfill
Promise.config({
  warnings: {
    wForgottenReturn: false,
  },
});

const polyfillFetch = () =>
  new Promise(resolve =>
    modernizr.on("fetch", result => {
      if (result) {
        return resolve();
      }

      return import(/* webpackChunkName: "fetch" */ "whatwg-fetch").then(
        resolve,
      );
    }),
  );

const polyfillURLSearchParams = () =>
  new Promise(resolve =>
    modernizr.on("urlsearchparams", result => {
      if (result) {
        return resolve();
      }

      return import(/* webpackChunkName: "url" */ "url-search-params-polyfill").then(
        resolve,
      );
    }),
  );

const polyfillIntl = () =>
  new Promise(resolve =>
    modernizr.on("intl", result => {
      if (result) {
        return resolve();
      }

      return Promise.all([
        import(/* webpackChunkName: "intl" */ "intl"),
        import(/* webpackChunkName: "intl" */ "intl/locale-data/jsonp/en.js"),
      ]).then(resolve);
    }),
  );

const polyfillIntlRelativeFormat = () =>
  import(/* webpackChunkName: "intl-relative-format" */ "intl-relativeformat").then(
    result => {
      window.IntlRelativeFormat = result.default;
      return import(/* webpackChunkName: "intl-relative-format" */ "intl-relativeformat/dist/locale-data/en.js");
    },
  );

export default () =>
  Promise.all([
    polyfillFetch(),
    polyfillURLSearchParams(),
    polyfillIntl(),
    polyfillIntlRelativeFormat(),
  ]).then(() => {});
