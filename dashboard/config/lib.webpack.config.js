import fs from "fs";
import path from "path";

import webpack from "webpack";
import CleanPlugin from "clean-webpack-plugin";

import makeConfig from "./base.webpack.config";

const root = fs.realpathSync(process.cwd());
const outputPath =
  process.env.NODE_ENV === "development"
    ? path.join(root, "build/lib-dev")
    : path.join(root, "build/lib");

const vendorPath = path.join(
  root,
  process.env.NODE_ENV === "development" ? "build/vendor-dev" : "build/vendor",
);

const libConfig = makeConfig({
  name: "lib",

  entry: {
    lib: [path.join(root, "src/lib")],
  },

  output: {
    path: path.join(outputPath, "public"),
    publicPath: "/",
    devtoolNamespace: "lib",
  },

  optimization: {
    // Disable "tree-shaking" by disabling es module export optimization.
    providedExports: false,
    usedExports: false,
  },

  plugins: [
    new CleanPlugin(outputPath, { root }),
    new webpack.DllReferencePlugin({
      manifest: path.join(vendorPath, "dll.json"),
    }),
  ],
});

libConfig.plugins.push(
  new webpack.DllPlugin({
    name: libConfig.output.library,
    path: path.join(outputPath, "dll.json"),
  }),
);

export default libConfig;
