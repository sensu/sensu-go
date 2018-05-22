import {
  NetworkError,
  ClientError,
  UnauthorizedError,
  ServerError,
} from "/errors/FetchError";

const doFetch = (path, config) =>
  // Wrap fetch call in bluebird promise to enable global rejection tracking
  Promise.resolve(fetch(path, config)).then(
    response => {
      if (response.status < 200) {
        throw new NetworkError(response.status, response.url, response);
      }

      if (response.status >= 500) {
        throw new ServerError(response);
      }

      if (response.status >= 400) {
        if (response.status === 401) {
          throw new UnauthorizedError(response);
        }
        throw new ClientError(response);
      }

      return response;
    },
    error => {
      throw new NetworkError(0, path, error);
    },
  );

export default doFetch;
