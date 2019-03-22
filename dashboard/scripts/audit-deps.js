import { execSync } from "child_process";
import chalk from "chalk";

// The severity levels that we can consider a passing result
const acceptable = ["info", "low"];

// Run report
let report;
try {
  report = execSync("yarn audit --json", { encoding: "utf-8" });
} catch (e) {
  report = e.stdout;
}
report = report.trim().split("\n");

// Collect results & dedup
const results = report.slice(0, -1).reduce((acc, raw) => {
  const response = JSON.parse(raw);
  const version = response.data.advisory.findings[0].version;
  const result = {
    ...response.data.advisory,
    version,
  };
  return { [`${result.module_name}${result.version}`]: result, ...acc };
}, {});

// Display advisories
Object.keys(results).forEach(key => {
  const result = results[key];
  console.info(
    chalk.blue("info"),
    "package",
    chalk.bold(`${result.module_name}@${result.version}`),
    "-",
    result.severity,
  );
});

// Tabulate deps with advisories outside of acceptable bounds
const unacceptableCount = Object.keys(results).reduce((acc, key) => {
  const result = results[key];
  return acceptable.indexOf(result.severity) > 0 ? acc : acc + 1;
}, 0);

// If any packages outside of bounds, report results and bomb
if (unacceptableCount > 0) {
  console.error(
    chalk.red("error"),
    chalk.bold(unacceptableCount),
    "moderate or higher severity vulnerabilities found.",
  );
  process.exit(1);
}

// Report totals
const total = Object.keys(results).length;
console.info(chalk.blue("info"), chalk.bold(total), "vulnerabilities found.");
process.exit(0);
