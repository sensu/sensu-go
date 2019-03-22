import "./util/exceptionHandler";

import jest from "jest";

process.env.NODE_ENV = process.env.NODE_ENV || "test";
const argv = process.argv.slice(3);

// Watch unless on CI or in coverage mode
if (!process.env.CI && argv.indexOf("--coverage") < 0) {
  argv.push("--watch");
}

jest.run(argv);
