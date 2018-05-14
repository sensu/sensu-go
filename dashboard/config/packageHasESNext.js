"use-strict";

const path = require("path");
const fs = require("fs");

const findRoot = require("find-root");

const cache = {};

module.exports = function packageHasESNext(filepath) {
  if (cache[filepath] !== undefined) {
    return cache[filepath];
  }

  const moduleRoot = findRoot(filepath);
  const packageFilepath = path.resolve(moduleRoot, "package.json");

  if (cache[packageFilepath] !== undefined) {
    return cache[packageFilepath];
  }

  const packageJsonText = fs.readFileSync(packageFilepath, {
    encoding: "utf-8",
  });

  const packageJson = JSON.parse(packageJsonText);
  const result = Object.prototype.hasOwnProperty.call(packageJson, "esnext");

  if (result) {
    cache[filepath] = result;
    cache[packageFilepath] = result;
  }

  return result;
};
