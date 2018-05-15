export default function(content) {
  const context = this;
  return content.default({
    ...context,
    content: JSON.stringify(content),
    emitFile(outputPath, fileContent) {
      context.emitFile(outputPath, fileContent);

      // return `module.exports = __webpack_public_path__ + ${JSON.stringify(
      //   outputPath,
      // )};

      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      // ¯\(°_o)/¯
      return `module.exports = ${fileContent}`;
    },
  });
}

export const raw = true;
