import gql from "graphql-tag";

const mutation = gql`
  mutation ExecuteCheckMutation($input: ExecuteCheckInput!) {
    executeCheck(input: $input) {
      errors {
        code
        input
      }
    }
  }
`;

export default (client, { id }) =>
  client
    .mutate({
      mutation,
      variables: {
        input: { id },
      },
    })
    .then(res => res.data.executeCheck);
