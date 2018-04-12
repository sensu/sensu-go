import { ApolloLink } from "apollo-link";
import { setContext } from "apollo-link-context";
import { HttpLink } from "apollo-link-http";
import { onError } from "apollo-link-error";

import {
  InMemoryCache,
  IntrospectionFragmentMatcher,
} from "apollo-cache-inmemory";
import ApolloClient from "apollo-client";

import { getAccessToken } from "../utils/authentication";

// TODO: Filter out out any type information unrelated to unions or interfaces
// prior to importing `schema.json`. IntrospectionFragmentMatcher only needs
// a subset of the schema.
// see: apollographql.com/docs/react/advanced/fragments.html#fragment-matcher
import { data as introspectionQueryResultData } from "../schema.json";

const fragmentMatcher = new IntrospectionFragmentMatcher({
  introspectionQueryResultData,
});

const cache = new InMemoryCache({
  fragmentMatcher,
  dataIdFromObject: object => object.id,
});

const authLink = setContext(() =>
  getAccessToken().then(token => ({
    headers: { Authorization: `Bearer ${token}` },
  })),
);

const errorLink = onError(({ graphQLErrors, networkError }) => {
  // TODO: Connect this error handler to display a blocking error alert
  if (graphQLErrors)
    graphQLErrors.forEach(error => {
      // eslint-disable-next-line no-console
      console.error(error.originalError || error);
    });
  // eslint-disable-next-line no-console
  if (networkError) console.error(networkError);
});

const httpLink = new HttpLink({
  uri: "/graphql",
  fetchOptions: {},
  credentials: "same-origin",
});

const client = new ApolloClient({
  cache,
  link: ApolloLink.from([errorLink, authLink, httpLink]),
});

export default client;
