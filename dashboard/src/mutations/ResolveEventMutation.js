import { graphql, commitMutation } from "react-relay";

const mutation = graphql`
  mutation ResolveEventMutation($input: ResolveEventInput!) {
    resolveEvent(input: $input) {
      event {
        timestamp
        check {
          status
          output
        }
      }
    }
  }
`;

function commit(environment, id) {
  return commitMutation(environment, {
    mutation,
    variables: {
      input: { id, source: "web app" },
    },
  });
}

export default { commit };
