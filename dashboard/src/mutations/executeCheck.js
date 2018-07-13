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

export default (client, { id, subscriptions = [] }) =>
  client
    .mutate({
      mutation,
      variables: {
        input: { id, subscriptions },
      },
    })
    .then(res => res.data.executeCheck);
