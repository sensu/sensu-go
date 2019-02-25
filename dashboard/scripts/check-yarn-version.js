import { execSync } from "child_process";
import chalk from "chalk";
import semver from "semver";

const requiredVersion = require("../package.json").engines.yarn;

const systemVersion = execSync("yarn --version", { encoding: "utf-8" }).trim();

if (!semver.satisfies(systemVersion, requiredVersion)) {
  console.error(
    chalk.red("error"),
    `System yarn ${chalk.bold(
      systemVersion,
    )} does not match required version ${chalk.bold(requiredVersion)}`,
  );
  console.log(
    chalk.blue("info"),
    "Visit",
    chalk.bold("https://yarnpkg.com/en/docs/install"),
    "for install and upgrade documentation.",
  );
  console.log();
  process.exit(1);
}
