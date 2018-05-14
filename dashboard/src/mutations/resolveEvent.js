import gql from "graphql-tag";

import handle from "/exceptionHandler";

const fragment = gql`
  fragment ResolveEventMutation_event on Event {
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

const source = "Web UI";
const getTime = () => {
  const now = new Date();
  return now.toUTCString();
};

export default (client, { id }) =>
  client.mutate({
    mutation,
    variables: {
      input: { id, source },
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
        handle(error);
      }
    },
    optimisticResponse: {
      resolveEvent: {
        event: {
          id,
          timestamp: getTime(),
          check: {
            status: 0,
            output: `Resolved manually with ${source}`,
            __typename: "Check",
          },
          __typename: "Event",
        },
        __typename: "ResolveEventPayload",
      },
      __typename: "Mutation",
    },
  });
