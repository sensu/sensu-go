import ExtendableError from "es6-error";

class QueryAbortedError extends ExtendableError {
  constructor(error) {
    super(error.name);
    this.original = error;
  }
}

export default QueryAbortedError;
