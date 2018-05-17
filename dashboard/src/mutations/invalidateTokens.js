import gql from "graphql-tag";

const mutation = gql`
  mutation InvalidateTokensMutation {
    invalidateTokens @client
  }
`;

export default client =>
  client.mutate({ mutation }).catch(error => {
    throw error.networkError || error;
  });
