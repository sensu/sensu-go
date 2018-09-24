import { ApolloLink, Observable } from "apollo-link";

import { when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";
import QueryAbortedError from "/errors/QueryAbortedError";

import refreshTokens from "/mutations/refreshTokens";
import flagTokens from "/mutations/flagTokens";

const EXPIRY_THRESHOLD_MS = 13 * 60 * 1000;
const MAX_REFRESHES = 3;

const authLink = ({ getClient }) =>
  new ApolloLink(
    (operation, forward) =>
      new Observable(observer => {
        let sub;

        const fetchToken = (attempts = 0) => {
          const forceRefresh = attempts > 0;

          refreshTokens(getClient(), {
            notBefore: forceRefresh
              ? null
              : new Date(Date.now() + EXPIRY_THRESHOLD_MS).toISOString(),
          })
            .then(
              ({ data }) => {
                const { auth } = data.refreshTokens;

                operation.setContext({
                  headers: {
                    Authorization: `Bearer ${auth.accessToken}`,
                  },
                });

                const nextObserver = {
                  next: (...args) => {
                    observer.next(...args);
                  },
                  complete: () => {
                    observer.complete();
                  },

                  // If chain results in an unauthorized error being thrown,
                  // either attempt to create a new access token or flag the
                  // auth token pair as invalid and throw query aborted err.
                  error: err => {
                    if (err instanceof UnauthorizedError) {
                      if (attempts < MAX_REFRESHES && !auth.invalid) {
                        sub.unsubscribe();
                        fetchToken(attempts + 1);
                      } else {
                        flagTokens(getClient());
                        observer.error(new QueryAbortedError(err));
                      }
                    } else {
                      observer.error(err);
                    }
                  },
                };
                sub = forward(operation).subscribe(nextObserver);
              },
              when(UnauthorizedError, error => {
                throw new QueryAbortedError(error);
              }),
            )
            .catch(observer.error.bind(observer));
        };

        fetchToken();

        return () => {
          if (sub) {
            sub.unsubscribe();
          }
        };
      }),
  );

export default authLink;
