import { ApolloLink } from "apollo-link";
import gql from "graphql-tag";

import {
  InMemoryCache,
  IntrospectionFragmentMatcher,
} from "apollo-cache-inmemory";
import ApolloClient from "apollo-client";

// https://www.apollographql.com/docs/react/advanced/fragments.html#fragment-matcher
import { data as introspectionQueryResultData } from "./schema/combinedTypes.macro";

import authLink from "./authLink";
import stateLink from "./stateLink";
import httpLink from "./httpLink";
import introspectionLink from "./introspectionLink";
import localStorageSync from "./localStorageSync";

const createClient = () => {
  const fragmentMatcher = new IntrospectionFragmentMatcher({
    introspectionQueryResultData,
  });

  const cache = new InMemoryCache({
    fragmentMatcher,
    dataIdFromObject: object => object.id,
  });

  let client = null;
  const getClient = () => {
    if (!client) {
      throw new Error("apollo client is not initialized");
    }
    return client;
  };

  client = new ApolloClient({
    cache,
    link: ApolloLink.from([
      introspectionLink(),
      stateLink({ cache }),
      authLink({ getClient }),
      httpLink(),
    ]),
  });

  localStorageSync(
    client,
    gql`
      query SyncAuthQuery {
        auth @client {
          accessToken
          refreshToken
          expiresAt
        }
      }
    `,
  );

  return client;
};

export default createClient;
