import { ApolloLink, Observable } from "apollo-link";

import { when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";
import QueryAbortedError from "/errors/QueryAbortedError";

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
              sub = forward(operation).subscribe(observer);
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
