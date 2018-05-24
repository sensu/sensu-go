const gitRevSync = __non_webpack_require__("git-rev-sync");
const fs = __non_webpack_require__("fs");
const path = __non_webpack_require__("path");

export default ({ rootContext }) => {
  const packageJson = JSON.parse(
    fs.readFileSync(path.join(rootContext, "package.json"), "utf8"),
  );
  const sourceRevision = gitRevSync.short();
  const sourceVersion = packageJson.version;
  const sourceURL = packageJson.repository.url;

  return `module.exports = ${JSON.stringify({
    sourceRevision,
    sourceVersion,
    sourceURL,
  })}`;
};
