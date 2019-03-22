import loaderUtils from "loader-utils";
import path from "path";

export default function(content) {
  const context = this;
  const options = loaderUtils.getOptions(context);

  return content.default({
    ...context,
    content: JSON.stringify(content),
    emitFile(relativePath, fileContent, emitOptions) {
      const currentOptions = {
        filename: "[name].[ext]",
        ...options,
        ...emitOptions,
      };

      const resourcePath = path.join(
        path.dirname(context.resourcePath),
        relativePath,
      );

      const filename = loaderUtils.interpolateName(
        { ...context, resourcePath },
        currentOptions.filename,
        { content: fileContent },
      );

      context.emitFile(filename, fileContent);

      return `module.exports = __webpack_public_path__ + ${JSON.stringify(
        filename,
      )}`;
    },
  });
}

export const raw = true;
