/* eslint-disable */
const path = require("path");
const script = path.resolve(path.dirname(__filename), process.argv[2]);
require("esm")(module)(script);
