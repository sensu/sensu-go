import gql from "graphql-tag";

import handle from "/exceptionHandler";

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
        handle(error);
      }
    },
  });
