import { ApolloLink, Observable } from "apollo-link";

import { when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";

import refreshTokens from "/mutations/refreshTokens";
import invalidateTokens from "/mutations/invalidateTokens";

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
              // Remove the current stored tokens if the token refresh attempt
              // fails. This will redirect the user back to the login screen.
              // We could potentially introduce a less disruptive behavior here,
              // maybe show a modal allowing the user to re-enter credentials
              // without loosing the current app state.
              invalidateTokens(getClient());

              // re-throw the UnauthorizedError instance to ensure later error
              // handling (likely a <Query> instance) can deal with it
              // appropriately.
              throw error;
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
