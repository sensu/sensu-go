import gql from "graphql-tag";

const fragment = gql`
  fragment ResolveEventMutation_event on Event {
    id
    timestamp
    check {
      status
      output
    }
  }
`;

const mutation = gql`
  mutation ResolveEventMutation($input: ResolveEventInput!) {
    resolveEvent(input: $input) {
      event {
        ...ResolveEventMutation_event
      }
    }
  }

  ${fragment}
`;

export default (client, { id }) =>
  client.mutate({
    mutation,
    variables: {
      input: { id, source: "Sensu web UI" },
    },
    update: (dataProxy, { data }) => {
      // Apollo wraps the call to `update` in its own try/catch and swallows any
      // errors. We must handle any errors ourselves if we don't want them to
      // be completely silenced.
      try {
        dataProxy.writeFragment({
          id,
          fragment,
          data: data.resolveEvent.event,
        });
      } catch (error) {
        // TODO: Connect this error handler to display a blocking error alert
        // eslint-disable-next-line no-console
        console.error(error);
      }
    },
  });
