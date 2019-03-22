import ExtendableError from "es6-error";

class ReactError extends ExtendableError {
  constructor(error, info) {
    const componentName = /^\n {4}in ([^\s]+)/.exec(info.componentStack)[1];

    super(
      `Error in <${componentName}> ${
        error.message ? ` (${error.message})` : ""
      }`,
    );

    this.stack = error.stack;
    this.componentStack = info.componentStack;

    this.original = error;
  }
}

export default ReactError;
