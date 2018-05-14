import chalk from "chalk";
import FileSizeReporter from "react-dev-utils/FileSizeReporter";
import formatWebpackMessages from "react-dev-utils/formatWebpackMessages";
import webpack from "webpack";

import "./exceptionHandler";
import assertEnv from "./assertEnv";
import getConfig from "../config/webpack.config";

process.env.NODE_ENV = process.env.NODE_ENV || "production";

assertEnv();

const WARN_AFTER_BUNDLE_GZIP_SIZE = 512 * 1024;
const WARN_AFTER_CHUNK_GZIP_SIZE = 1024 * 1024;

const config = getConfig();

FileSizeReporter.measureFileSizesBeforeBuild(config.output.path).then(
  previousFileSizes => {
    console.log(`Starting webpack compiler in ${config.mode} mode`);

    const compiler = webpack(config);

    compiler.run((error, stats) => {
      if (error) {
        throw error;
      }

      const messages = formatWebpackMessages(stats.toJson({}, true));

      if (messages.errors.length) {
        throw new Error(messages.errors[0]);
      }

      if (messages.warnings.length) {
        console.log(chalk.yellow("Compiled with warnings."));
        console.log();
        console.log(messages.warnings.join("\n\n"));
      } else {
        console.log(chalk.green("Compiled successfully."));
      }

      console.log();
      console.log("File sizes after gzip:");
      console.log();

      FileSizeReporter.printFileSizesAfterBuild(
        stats,
        previousFileSizes,
        config.output.path,
        WARN_AFTER_BUNDLE_GZIP_SIZE,
        WARN_AFTER_CHUNK_GZIP_SIZE,
      );
      console.log();
    });
  },
);
