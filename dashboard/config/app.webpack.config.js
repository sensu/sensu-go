import fs from "fs";
import path from "path";
import webpack from "webpack";
import HtmlWebpackPlugin from "html-webpack-plugin";
import AddAssetHtmlPlugin from "add-asset-html-webpack-plugin";

import CleanPlugin from "clean-webpack-plugin";

import makeConfig from "./base.webpack.config";

const root = fs.realpathSync(process.cwd());
const outputPath = path.join(root, "build/app");

const getBundleAssets = (stats, chunk) =>
  []
    .concat(stats.assetsByChunkName[chunk])
    .filter(name => /\.js$/.test(name))
    .map(name => ({
      filepath: path.join(root, stats.outputPath, name),
      outputPath: path.join(".", "static", "js"),
      publicPath: path.join("/", "static", "js"),
    }));

export default makeConfig({
  name: "app",

  entry: {
    app: [path.join(root, "src/app")],
  },

  output: {
    path: path.join(outputPath, "public"),
    publicPath: "/",
    devtoolNamespace: "app",
  },

  plugins: [
    new CleanPlugin(outputPath, { root }),

    // Generates an `index.html` file with the <script> injected.
    new HtmlWebpackPlugin({
      inject: true,
      template: path.resolve(root, "src/app/static/index.html"),
      minify: process.env.NODE_ENV !== "development" && {
        removeComments: true,
        collapseWhitespace: true,
        removeRedundantAttributes: true,
        useShortDoctype: true,
        removeEmptyAttributes: true,
        removeStyleLinkTypeAttributes: true,
        keepClosingSlash: true,
        minifyJS: true,
        minifyCSS: true,
        minifyURLs: true,
      },
    }),

    new webpack.DllReferencePlugin({
      manifest: path.join(root, "build/vendor/dll.json"),
    }),
    new AddAssetHtmlPlugin([
      // eslint-disable-next-line import/no-unresolved
      ...getBundleAssets(require("../build/vendor/stats.json"), "vendor"),
    ]),
  ].concat(
    process.env.NODE_ENV !== "development" && [
      new webpack.DllReferencePlugin({
        manifest: path.join(root, "build/lib/dll.json"),
      }),
      new AddAssetHtmlPlugin([
        // eslint-disable-next-line import/no-unresolved
        ...getBundleAssets(require("../build/lib/stats.json"), "vendor"),
      ]),
    ],
  ),
});
