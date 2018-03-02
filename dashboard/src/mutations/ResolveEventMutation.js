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

function commit(environment, id, { onCompleted }) {
  return commitMutation(environment, {
    mutation,
    onCompleted,
    variables: {
      input: { id, source: "Sensu web UI" },
    },
    updater: (store, data) => {
      const result = data.resolveEvent.event;
      const ev = store.get(id);
      ev.setValue(result.timestamp, "timestamp");

      const check = ev.getLinkedRecord("check");
      check.setValue(result.check.output, "output");
      check.setValue(result.check.status, "status");
    },
  });
}

export default { commit };
