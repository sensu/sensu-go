import { BatchHttpLink as HttpLink } from "apollo-link-batch-http";
import doFetch from "/utils/fetch";
import { FailedError } from "/errors/FetchError";
import gql from "graphql-tag";

const mutation = gql`
  mutation SetLocalNetworkOfflineMutation($offline: Boolean!) {
    setLocalNetworkOffline(offline: $offline) @client
  }
`;

const httpLink = ({ getClient }) =>
  new HttpLink({
    uri: "/graphql",
    fetch: (url, init) =>
      doFetch(url, init).then(
        response => {
          getClient().mutate({
            mutation,
            variables: { offline: false },
          });
          return response;
        },
        error => {
          getClient().mutate({
            mutation,
            variables: { offline: error instanceof FailedError },
          });

          throw error;
        },
      ),
    credentials: "same-origin",
    batchMax: 25,
    batchInterval: 3,
  });

export default httpLink;
