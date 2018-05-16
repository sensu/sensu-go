import { ApolloLink } from "apollo-link";
import gql from "graphql-tag";

import {
  InMemoryCache,
  IntrospectionFragmentMatcher,
} from "apollo-cache-inmemory";
import ApolloClient from "apollo-client";

// TODO: Filter out out any type information unrelated to unions or interfaces
// prior to importing `schema.json`. IntrospectionFragmentMatcher only needs
// a subset of the schema.
// see: apollographql.com/docs/react/advanced/fragments.html#fragment-matcher
import { data as introspectionQueryResultData } from "/schema.json";

import authLink from "./authLink";
import stateLink from "./stateLink";
import httpLink from "./httpLink";
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
    if (!client) throw new Error("apollo client is not initialized");
    return client;
  };

  client = new ApolloClient({
    cache,
    link: ApolloLink.from([
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
