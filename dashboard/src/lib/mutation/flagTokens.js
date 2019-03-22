import gql from "graphql-tag";

const mutation = gql`
  mutation FlagTokensMutation {
    flagTokens @client
  }
`;

export default client => client.mutate({ mutation });
