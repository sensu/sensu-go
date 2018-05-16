import capture from "bugnet";
import ErrorStackParser from "error-stack-parser";

let a;
const absolute = path => {
  a = a || document.createElement("a");
  a.href = path;
  return a.href;
};

const publicPath = absolute(__webpack_public_path__);

let lastError;

const handle = error => {
  if (!error || lastError === error) {
    return;
  }

  lastError = error;

  let frames;

  try {
    frames = ErrorStackParser.parse(error);
  } catch (e) {
    frames = [];
  }

  // Detect that the caught error originated from our own code. If the source
  // file name parsed from the error stack is not under the asset public path,
  // consider the error as originating from a vendor script.
  if (
    frames[0] !== undefined &&
    typeof frames[0].fileName === "string" &&
    frames[0].fileName.indexOf(publicPath) !== 0
  ) {
    // Ignore vendor script error.
  } else {
    // eslint-disable-next-line no-console
    console.log("exceptionHandler.js", error);
  }
};

capture(handle);

export default handle;
