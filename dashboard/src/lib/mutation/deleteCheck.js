import gql from "graphql-tag";
import handle from "/lib/exceptionHandler";

const fragment = gql`
  fragment DeleteCheckMutation_check on CheckConfig {
    deleted @client
  }
`;

const mutation = gql`
  mutation DeleteCheckMutation($input: DeleteRecordInput!) {
    deleteCheck(input: $input) {
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
    optimisticResponse: {
      deleteCheck: {
        deletedId: id,
        __typename: "DeleteRecordPayload",
      },
      __typename: "Mutation",
    },
  });
