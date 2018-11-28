// @flow

import ExtendableError from "es6-error";

class QueryAbortedError extends ExtendableError {
  original: Error | null;

  constructor(error: Error) {
    super(error.name);
    this.original = error;
  }
}

export default QueryAbortedError;
