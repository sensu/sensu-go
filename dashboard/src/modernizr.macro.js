const detectFeatures = [
  "es6/collections",
  "network/fetch",
  "url/urlsearchparams",
  "intl",
];

const minify = process.env.NODE_ENV === "production";

export default ({ async, emitFile }) => {
  const done = async();

  // Create a custom build of Modernizr for the features we need to detect.
  __non_webpack_require__("modernizr").build(
    { minify, "feature-detects": detectFeatures },
    output => {
      // Convert the generated bundle to a CJS module that exports Modernizer
      // removing any global side-effects.
      const patched = minify
        ? output
            .replace(/^!(function\()/m, "module.exports=$1")
            .replace(/;[a-z]+\.Modernizr=Modernizr/, ";return Modernizr")
        : output
            .replace(/^;(\(function\()/m, "module.exports=$1")
            .replace(/window.Modernizr = Modernizr;$/m, "return Modernizr");

      emitFile("modernizr.js", patched);
      done(null, patched);
    },
  );
};
