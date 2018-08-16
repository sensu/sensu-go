import { ApolloLink, Observable } from "apollo-link";

import { when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";
import QueryAbortedError from "/errors/QueryAbortedError";

import flagTokens from "/mutations/flagTokens";
import refreshTokens from "/mutations/refreshTokens";

const EXPIRY_THRESHOLD_MS = 13 * 60 * 1000;

const authLink = ({ getClient }) =>
  new ApolloLink(
    (operation, forward) =>
      new Observable(observer => {
        let sub;
        refreshTokens(getClient(), {
          notBefore: new Date(Date.now() + EXPIRY_THRESHOLD_MS).toISOString(),
        })
          .then(
            ({ data }) => {
              operation.setContext({
                headers: {
                  Authorization: `Bearer ${
                    data.refreshTokens.auth.accessToken
                  }`,
                },
              });

              const nextObserver = {
                next: observer.next.bind(observer),
                complete: observer.complete.bind(observer),

                // If chain results in an unauthorized error being thrown,
                // flag the auth token pair as invalid and throw aborted err.
                error: err => {
                  if (err instanceof UnauthorizedError) {
                    flagTokens(getClient());
                    observer.error(new QueryAbortedError(err));
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

        return () => {
          if (sub) {
            sub.unsubscribe();
          }
        };
      }),
  );

export default authLink;
