// @flow

import {
  FailedError,
  ClientError,
  UnauthorizedError,
  ServerError,
} from "/errors/FetchError";

const doFetch: typeof fetch = (input, config) =>
  // Wrap fetch call in bluebird promise to enable global rejection tracking
  Promise.resolve(fetch(input, config)).then(
    response => {
      if (response.status < 200) {
        throw new FailedError(response.status, input, response);
      }

      if (response.status >= 500) {
        throw new ServerError(response.status, input, response);
      }

      if (response.status >= 400) {
        if (response.status === 401) {
          throw new UnauthorizedError(response.status, input, response);
        }
        throw new ClientError(response.status, input, response);
      }

      return response;
    },
    error => {
      throw new FailedError(0, input, null, error);
    },
  );

export default doFetch;
