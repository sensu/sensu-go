import chalk from "chalk";

const logAndDie = error => {
  if (error && error.message) {
    console.error(chalk.red(`${error.name}: ${error.message}`));
    if (error.stack) {
      const frames = error.stack.split("\n");
      const skip = frames.findIndex(frame => /^ {4}at/.test(frame));
      console.error(
        chalk.gray(
          `${skip !== -1 ? frames.slice(skip).join("\n") : error.stack}\n`,
        ),
      );
    }
  } else {
    console.error(error);
  }
  process.exit(1);
};

process.on("unhandledRejection", logAndDie);
process.on("uncaughtException", logAndDie);
