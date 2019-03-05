import gql from "graphql-tag";

const mutation = gql`
  mutation RetryLocalNetworkMutation {
    retryLocalNetwork @client
  }
`;

export default client => client.mutate({ mutation });
