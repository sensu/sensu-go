import fs from "fs";
import path from "path";
import webpack from "webpack";
import HtmlWebpackPlugin from "html-webpack-plugin";
import SWPrecacheWebpackPlugin from "sw-precache-webpack-plugin";
import eslintFormatter from "react-dev-utils/eslintFormatter";
import CaseSensitivePathsPlugin from "case-sensitive-paths-webpack-plugin";
import CleanPlugin from "clean-webpack-plugin";
import { StatsWriterPlugin } from "webpack-stats-plugin";

export default () => {
  // Make sure any symlinks in the project folder are resolved:
  // https://github.com/facebookincubator/create-react-app/issues/637
  const root = fs.realpathSync(process.cwd());

  const isDevelopment = process.env.NODE_ENV === "development";
  const isProduction = process.env.NODE_ENV === "production";

  let devtool = false;

  if (process.env.GENERATE_SOURCEMAP === "true") {
    devtool = "source-map";
  } else if (!isProduction) {
    devtool = "cheap-module-source-map";
  }

  const outputPath = path.resolve(root, "build");

  return {
    bail: true,
    mode: process.env.NODE_ENV,

    devtool,

    entry: [path.resolve(root, "src/index.js")],

    output: {
      path: outputPath,

      publicPath: "/",

      pathinfo: isDevelopment,

      filename: isProduction
        ? "static/js/[name].[hash:8].js"
        : "static/js/[name].js",

      chunkFilename: isProduction
        ? "static/js/[name].[chunkhash:8].chunk.js"
        : "static/js/[name].chunk.js",

      // Point sourcemap entries to original disk location (format as URL on Windows)
      devtoolModuleFilenameTemplate: ({ absoluteResourcePath }) =>
        path
          .relative(path.resolve(root, "src"), absoluteResourcePath)
          .replace(/\\/g, "/"),
    },

    optimization: {
      splitChunks: { minChunks: 2 },
    },

    resolve: {
      extensions: [".web.js", ".js", ".json", ".web.jsx", ".jsx"],
      alias: {
        // Alias any reference to babel runtime Promise to bluebird. This
        // prevents duplicate promise polyfills in the build.
        "babel-runtime/core-js/promise": "bluebird/js/browser/bluebird.core.js",
      },
    },

    module: {
      strictExportPresence: true,
      rules: [
        {
          enforce: "pre",
          test: /\.(js|jsx)$/,
          exclude: path.resolve(root, "node_modules"),
          use: [
            {
              loader: require.resolve("eslint-loader"),
              options: {
                formatter: eslintFormatter,
                eslintPath: require.resolve("eslint"),
                emitError: false,
                emitWarning: isDevelopment,
              },
            },
          ],
        },
        {
          oneOf: [
            {
              test: [/\.bmp$/, /\.gif$/, /\.jpe?g$/, /\.png$/],
              loader: require.resolve("url-loader"),
              options: {
                limit: 10000,
                name: "static/media/[name].[hash:8].[ext]",
              },
            },
            {
              test: /\.macro\.js$/,
              exclude: path.resolve(root, "node_modules"),
              loaders: [
                require.resolve("./macroLoader"),
                require.resolve("value-loader"),
              ],
            },
            {
              test: /\.(js|jsx)$/,
              exclude: path.resolve(root, "node_modules"),
              loader: require.resolve("babel-loader"),
              options: {
                compact: isProduction,
                cacheDirectory: isDevelopment,
              },
            },
            {
              loader: require.resolve("file-loader"),
              exclude: [/\.js$/, /\.html$/, /\.json$/],
              options: {
                name: "static/media/[name].[hash:8].[ext]",
              },
            },
          ],
        },
      ],
    },
    plugins: [
      new StatsWriterPlugin({
        filename: "../stats.json",
        fields: [
          "version",
          "hash",
          "time",
          "builtAt",
          "assetsByChunkName",
          "assets",
          "filteredAssets",
          "entrypoints",
          "modules",
          "filteredModules",
          "children",
        ],
      }),
      new CleanPlugin([outputPath, path.resolve(outputPath, "../stats.json")], {
        root,
        verbose: false,
      }),
      new webpack.DefinePlugin({
        NODE_ENV: JSON.stringify(process.env.NODE_ENV),
      }),
      new webpack.ProvidePlugin({
        // Alias any reference to global Promise object to bluebird.
        Promise: require.resolve("bluebird/js/browser/bluebird.core.js"),
      }),
      // Generates an `index.html` file with the <script> injected.
      new HtmlWebpackPlugin({
        inject: true,
        template: path.resolve(root, "src/static/index.html"),
        minify: isProduction && {
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
      // Remove moment locales.
      new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
      new webpack.HashedModuleIdsPlugin(),
    ]
      .concat(
        isDevelopment && [
          new webpack.NamedModulesPlugin(),
          new webpack.HotModuleReplacementPlugin(),
          new CaseSensitivePathsPlugin(),
        ],
      )
      .concat(
        isProduction && [
          new SWPrecacheWebpackPlugin({
            // If a URL is already hashed by Webpack, then there is no concern
            // about it being stale, and the cache-busting can be skipped.
            dontCacheBustUrlsMatching: /\.\w{8}\./,
            filename: "service-worker.js",
            minify: true,
            logger: () => {},
            // For unknown URLs, fallback to the index page
            navigateFallback: `/index.html`,
            // Don't precache sourcemaps (they're large) and build asset manifest:
            staticFileGlobsIgnorePatterns: [/\.map$/],
          }),
        ],
      )
      .filter(Boolean),
  };
};
