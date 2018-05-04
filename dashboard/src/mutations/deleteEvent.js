import gql from "graphql-tag";

const fragment = gql`
  fragment DeleteEventMutation_event on Event {
    deleted @client
  }
`;

const mutation = gql`
  mutation DeleteEventMutation($input: DeleteRecordInput!) {
    deleteEvent(input: $input) {
      deletedId
    }
  }
`;

export default (client, { id }) =>
  client.mutate({
    mutation,
    variables: {
      input: { id },
    },
    update: cache => {
      try {
        const ev = cache.readFragment({ fragment, id });
        const data = { ...ev, deleted: true };
        cache.writeFragment({ fragment, id, data });
      } catch (error) {
        // TODO: Connect this error handler to display a blocking error alert
        // eslint-disable-next-line no-console
        console.error(error);
      }
    },
  });
