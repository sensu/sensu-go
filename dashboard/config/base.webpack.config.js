import fs from "fs";
import path from "path";
import webpack from "webpack";
import CaseSensitivePathsPlugin from "case-sensitive-paths-webpack-plugin";
import TerserPlugin from "terser-webpack-plugin";
import { StatsWriterPlugin } from "webpack-stats-plugin";

const root = fs.realpathSync(process.cwd());

const jsPath = path.join("static", "js");
const mediaPath = path.join("static", "media");

export default ({
  // Include a hash of each file's content in its name unless running a
  // development build. This ensures browser caches are automatically invalided
  // by newer asset versions. We avoid using [hash] here since it represents
  // a hash of the entire build, and not of each individual file. Using it would
  // cause a update to any file in the build to change the name of all files,
  // even ones that that didn't change from the previous build.
  contentHashName = process.env.NODE_ENV === "development"
    ? "[name]"
    : "[name]_[contenthash:4]",
  chunkHashName = process.env.NODE_ENV === "development"
    ? "[name]"
    : "[name]_[chunkhash:4]",

  // file-loader calculates hashes differently from the rest of webpack
  // [hash] in file loader is equivalent to [contenthash] elsewhere
  // see: https://github.com/webpack-contrib/file-loader#placeholders
  fileLoaderHashName = process.env.NODE_ENV === "development"
    ? "[name]"
    : "[name]_[hash:4]",

  context,
  entry,
  output,
  devServer,
  plugins,
  optimization,
  module: { rules = [], ...module } = {},
  target = "web",
  name,
  ...config
}) => ({
  name,
  context: root,
  bail: true,
  target,
  mode: process.env.NODE_ENV,
  entry,

  devtool: process.env.NODE_ENV === "development" ? "source-map" : false,

  output: {
    filename: path.join(jsPath, `${contentHashName}.js`),
    chunkFilename: path.join(jsPath, `${chunkHashName}.js`),
    library: chunkHashName,
    pathinfo: process.env.NODE_ENV === "development",
    ...output,
  },

  optimization: {
    splitChunks: { minChunks: 2 },
    minimizer: [
      new TerserPlugin({
        sourceMap: true,
        terserOptions: {
          // Disable function name minification in order to preserve class
          // names. This makes tracking down bugs in production builds far
          // more manageable at the expense of slightly larger (about 15%)
          // compressed bundle sizes.
          keep_fnames: true,
        },
      }),
    ],
    ...optimization,
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
        oneOf: [
          {
            test: /\.css$/,
            use: ["style-loader", "css-loader"],
          },
          {
            test: /\.worker\.js$/,
            loader: "worker-loader",
            options: {
              name: path.join(jsPath, `${contentHashName}.js`),
            },
          },
          {
            test: /\.macro\.js$/,
            exclude: path.resolve(root, "node_modules"),
            loaders: [
              {
                loader: require.resolve("./macroLoader"),
                options: {
                  filename: path.join(mediaPath, `${fileLoaderHashName}.[ext]`),
                },
              },
              // Value-loader is the source of any Tapable deprecation
              // warnings.
              require.resolve("value-loader"),
            ],
          },
          {
            test: /\.(js|jsx)$/,
            exclude: path.resolve(root, "node_modules"),
            loader: require.resolve("babel-loader"),
            options: {
              cacheDirectory: process.env.NODE_ENV === "development",
            },
          },
          {
            loader: require.resolve("file-loader"),
            exclude: [/\.js$/, /\.html$/, /\.json$/],
            options: {
              name: path.join(mediaPath, `${fileLoaderHashName}.[ext]`),
            },
          },
          {
            test: /\.html$/,
            loader: require.resolve("html-loader"),
            options: {
              interpolate: true,
            },
          },
        ],
      },
      ...rules,
    ],
    ...module,
  },

  plugins: [
    new StatsWriterPlugin({
      filename: "../stats.json",
      fields: null,
      transform(stats) {
        return JSON.stringify(
          {
            version: stats.version,
            hash: stats.hash,
            outputPath: path.relative(root, stats.outputPath),
            assetsByChunkName: stats.assetsByChunkName,
            assets: stats.assets,
            chunks: stats.chunks.map(chunk => ({
              id: chunk.id,
              rendered: chunk.rendered,
              initial: chunk.initial,
              entry: chunk.entry,
              size: chunk.size,
              names: chunk.names,
              files: chunk.files,
              hash: chunk.hash,
            })),
          },
          null,
          2,
        );
      },
    }),
    new webpack.ProvidePlugin({
      // Alias any reference to global Promise object to bluebird.
      Promise: require.resolve("bluebird/js/browser/bluebird.core.js"),
    }),
    // Prevent moment locales from getting bundled.
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
  ]
    .concat(
      process.env.NODE_ENV !== "development" && [
        // Use hashed module ids in production builds.
        new webpack.HashedModuleIdsPlugin(),
      ],
    )
    .concat(
      process.env.NODE_ENV === "development" && [
        new CaseSensitivePathsPlugin(),
      ],
    )
    .concat(plugins)
    .filter(Boolean),

  ...config,
});
