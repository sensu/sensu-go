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

export default (client, args) =>
  client
    .mutate({
      mutation,
      variables: {
        input: args,
      },
    })
    .then(res => res.data.executeCheck);
