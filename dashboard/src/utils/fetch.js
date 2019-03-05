// @flow
import type { ApolloCache } from "react-apollo";

import {
  FailedError,
  ClientError,
  UnauthorizedError,
  ServerError,
} from "/errors/FetchError";

import { setOffline } from "/apollo/resolvers/localNetwork";

const doFetch = (cache: ApolloCache<mixed>): typeof fetch => (input, config) =>
  // Wrap fetch call in bluebird promise to enable global rejection tracking
  Promise.resolve(fetch(input, config)).then(
    response => {
      if (response.status === 0) {
        // The request failed for one of a number of possible reasons:
        //  - blocked by CORS
        //  - blocked by a content blocker browser extension
        //  - blocked by a firewall
        //
        // Unfortunately the specific case is not made available. (I suspect the
        // browser doesn't provide more detail in order to protect the user from
        // potentially harmful scripts. It's worth noting that specific details
        // are often logged to the dev tools console).
        //
        //  For our purposes we'll consider that all cases reflect the offline
        //  state in the app.e
        setOffline(cache, true);

        throw new FailedError(response.status, input, response);
      }

      // Clear the offline flag in the apollo cache
      setOffline(cache, false);

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
      // Set the offline flag in the apollo cache
      setOffline(cache, true);
      throw new FailedError(0, input, null, error);
    },
  );

export default doFetch;
