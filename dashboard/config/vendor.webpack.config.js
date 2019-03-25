import fs from "fs";
import path from "path";

import webpack from "webpack";
import CleanPlugin from "clean-webpack-plugin";

import makeConfig from "./base.webpack.config";

const root = fs.realpathSync(process.cwd());
const outputPath =
  process.env.NODE_ENV === "development"
    ? path.join(root, "build/vendor-dev")
    : path.join(root, "build/vendor");

const vendorConfig = makeConfig({
  name: "vendor",

  entry: {
    vendor: [
      "react",
      "react-dom",
      "react-router-dom",
      "graphql-tag",
      "react-apollo",
      "@material-ui/core",
      "react-resize-observer",
      "prop-types",
      "classnames",
      "react-spring",
      "fbjs",
    ],
  },

  output: {
    path: path.join(outputPath, "public"),
    publicPath: "/",
    devtoolNamespace: "vendor",
  },

  optimization: {
    // Disable "tree-shaking" by disabling es module export optimization.
    providedExports: false,
    usedExports: false,
  },

  plugins: [new CleanPlugin(outputPath, { root })],
});

vendorConfig.plugins.push(
  new webpack.DllPlugin({
    name: vendorConfig.output.library,
    path: path.join(outputPath, "dll.json"),
  }),
);

export default vendorConfig;
