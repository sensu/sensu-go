import gql from "graphql-tag";

const mutation = gql`
  mutation CreateTokensMutation($username: String!, $password: String!) {
    createTokens(username: $username, password: $password) @client
  }
`;

export default (client, { username, password } = {}) =>
  client
    .mutate({ mutation, variables: { username, password } })
    .catch(error => {
      throw error.networkError || error;
    });
