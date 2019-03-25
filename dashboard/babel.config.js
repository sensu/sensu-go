const path = require("path");

module.exports = {
  presets: [
    [
      "@babel/preset-env",
      {
        targets: {
          ie: "11",
        },
        // Disable polyfill transforms.
        useBuiltIns: false,
        // Do not transform es6 modules, required for webpack "tree shaking".
        modules: false,
      },
    ],
    "@babel/preset-react",
    "@babel/preset-flow",
  ],
  plugins: [
    "@babel/plugin-syntax-dynamic-import",
    "@babel/plugin-transform-destructuring",
    [
      "@babel/plugin-proposal-class-properties",
      {
        loose: true,
      },
    ],
    [
      "@babel/plugin-proposal-object-rest-spread",
      {
        useBuiltIns: true,
      },
    ],
    [
      "@babel/plugin-transform-react-jsx",
      {
        useBuiltIns: true,
      },
    ],
    [
      "module-resolver",
      {
        alias: {
          "": path.join(__dirname, "src"),
        },
      },
    ],
  ],
  env: {
    development: {
      plugins: [
        // "@babel/plugin-transform-react-jsx-source",
        // "@babel/plugin-transform-react-jsx-self"
      ],
    },
    test: {
      presets: [
        [
          "@babel/preset-env",
          {
            targets: {
              node: "current",
            },
            // Disable polyfill transforms.
            useBuiltIns: false,
            // Transform modules to CJS for node runtime.
            modules: "commonjs",
          },
        ],
      ],
    },
    production: {
      plugins: [
        [
          "transform-react-remove-prop-types",
          {
            removeImport: true,
          },
        ],
      ],
    },
  },
};
