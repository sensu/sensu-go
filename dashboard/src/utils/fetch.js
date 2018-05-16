import {
  NetworkError,
  ClientError,
  UnauthorizedError,
  ServerError,
} from "/errors/FetchError";

const doFetch = async (path, config) => {
  const response = await fetch(path, config).catch(error => {
    throw new NetworkError(0, path, error);
  });

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
};

export default doFetch;
