import gql from "graphql-tag";

const mutation = gql`
  mutation RefreshTokensMutation($notBefore: DateTime) {
    refreshTokens(notBefore: $notBefore) @client {
      auth {
        __typename
        accessToken
        refreshToken
        expiresAt
      }
    }
  }
`;

export default (client, { notBefore } = {}) =>
  client.mutate({ mutation, variables: { notBefore } });
