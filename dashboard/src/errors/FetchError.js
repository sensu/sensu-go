import ExtendableError from "es6-error";

export default class FetchError extends ExtendableError {
  constructor(status, url, response, error) {
    super(`${response.status}`);

    if (this.constructor === FetchError) {
      throw new TypeError("Can't initiate an abstract class.");
    }

    this.status = status;
    this.url = url;
    this.response = response;
    this.original = error;
  }
}

export class ParseError extends FetchError {
  constructor(response, error) {
    super(response.status, response.url, undefined, error);
  }
}

export class NetworkError extends FetchError {}

export class ServerError extends FetchError {
  constructor(response) {
    super(response.status, response.url, response);
  }
}

export class ClientError extends FetchError {
  constructor(response) {
    super(response.status, response.url, response);
  }
}

export class UnauthorizedError extends ClientError {}
