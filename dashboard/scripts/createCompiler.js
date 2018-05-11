import chalk from "chalk";
import path from "path";
import formatWebpackMessages from "react-dev-utils/formatWebpackMessages";
import webpack from "webpack";

import { success, loading } from "./log";

export default config => {
  loading.start(`Starting webpack compiler in ${config.mode} mode`);

  const compiler = webpack(config);

  compiler.hooks.failed.tap("createCompiler", error => {
    throw error;
  });

  compiler.hooks.invalid.tap("createCompiler", file => {
    console.log();
    loading.start(
      `Compiling ${chalk.gray(path.relative(process.cwd(), file))}`,
    );
  });

  compiler.hooks.done.tap("createCompiler", stats => {
    loading.stop();
    const messages = formatWebpackMessages(stats.toJson({}, true));

    if (!messages.errors.length && !messages.warnings.length) {
      console.log();
      console.log(success(), chalk.green("Compiled successfully!"));
    }

    if (messages.errors.length) {
      if (config.mode === "development") {
        console.log();
        console.log(chalk.red("Failed to compile."));
        console.log();
        console.log(messages.errors.join("\n\n"));
      } else {
        setImmediate(() => {
          throw new Error(messages.errors.join("\n\n"));
        });
      }
    } else if (messages.warnings.length) {
      if (config.mode === "development") {
        console.log();
        console.log(chalk.yellow("Compiled with warnings."));
        console.log();
        console.log(messages.warnings.join("\n\n"));
      }
    }
  });

  return compiler;
};
