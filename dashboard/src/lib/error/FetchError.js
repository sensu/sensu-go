// @flow

import ExtendableError from "es6-error";

export default class FetchError extends ExtendableError {
  statusCode: number;
  input: RequestInfo;
  original: Error | null;
  response: Response | null;

  constructor(
    status: number,
    input: RequestInfo,
    response: Response | null = null,
    error: Error | null = null,
  ) {
    super(`${status}`);

    if (this.constructor === FetchError) {
      throw new TypeError("Can't initiate an abstract class.");
    }

    this.statusCode = status;
    this.input = input;
    this.response = response;
    this.original = error;
  }
}

export class FailedError extends FetchError {}

export class ServerError extends FetchError {
  constructor(status: number, input: RequestInfo, response: ?Response) {
    super(status, input, response, null);
  }
}

export class ClientError extends FetchError {
  constructor(status: number, input: RequestInfo, response: ?Response) {
    super(status, input, response, null);
  }
}

export class UnauthorizedError extends ClientError {}
