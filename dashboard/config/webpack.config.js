import fs from "fs";
import path from "path";
import webpack from "webpack";
import HtmlWebpackPlugin from "html-webpack-plugin";
import eslintFormatter from "react-dev-utils/eslintFormatter";
import CaseSensitivePathsPlugin from "case-sensitive-paths-webpack-plugin";
import CleanPlugin from "clean-webpack-plugin";
import { StatsWriterPlugin } from "webpack-stats-plugin";
import UglifyJsPlugin from "uglifyjs-webpack-plugin";

export default () => {
  // Make sure any symlinks in the project folder are resolved:
  // https://github.com/facebookincubator/create-react-app/issues/637
  const root = fs.realpathSync(process.cwd());

  const isDevelopment = process.env.NODE_ENV === "development";
  const isProduction = process.env.NODE_ENV === "production";

  const outputPath = path.resolve(root, "build");

  return {
    bail: true,
    mode: process.env.NODE_ENV,

    devtool: "source-map",

    entry: [path.resolve(root, "src/index.js")],

    output: {
      globalObject: "self",

      path: outputPath,

      publicPath: "/",

      pathinfo: isDevelopment,

      filename: isProduction
        ? "static/js/[name].[hash:8].js"
        : "static/js/[name].js",

      chunkFilename: isProduction
        ? "static/js/[name].[chunkhash:8].chunk.js"
        : "static/js/[name].chunk.js",

      devtoolModuleFilenameTemplate: ({ absoluteResourcePath }) =>
        path
          .relative(path.resolve(root), absoluteResourcePath)
          .replace(/\\/g, "/"),
    },

    optimization: {
      splitChunks: { minChunks: 2 },
      minimizer: [
        new UglifyJsPlugin({
          sourceMap: true,
          uglifyOptions: {
            // Disable function name minification in order to preserve class
            // names. This makes tracking down bugs in production builds far
            // more manageable.
            keep_fnames: true,
          },
        }),
      ],
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
          enforce: "pre",
          test: /\.js$/,
          include: path.resolve(root, "node_modules"),
          exclude: [path.resolve(root, "node_modules/apollo-client")],
          loader: require.resolve("source-map-loader"),
          options: {
            includeModulePaths: true,
          },
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
              test: /\.css$/,
              use: ["style-loader", "css-loader"],
            },
            {
              test: /\.worker\.js$/,
              loader: "worker-loader",
              options: {
                name: "static/[hash].worker.js",
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
            {
              test: /\.html$/,
              loader: require.resolve("raw-loader"),
            },
          ],
        },
      ],
    },
    plugins: [
      new StatsWriterPlugin({
        filename: "../stats.json",
        fields: [
          "assets",
          "assetsByChunkName",
          "builtAt",
          "children",
          "chunks",
          "entrypoints",
          "errors",
          "filteredAssets",
          "filteredModules",
          "hash",
          "modules",
          "time",
          "version",
          "warnings",
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
      .filter(Boolean),
  };
};
