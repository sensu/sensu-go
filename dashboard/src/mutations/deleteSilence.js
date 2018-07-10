import gql from "graphql-tag";

const fragment = gql`
  fragment DeleteSilenceMutation_silence on Silenced {
    deleted @client
  }
`;

const mutation = gql`
  mutation DeleteSilenceMutation($input: DeleteRecordInput!) {
    deleteSilence(input: $input) {
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
      const ev = cache.readFragment({ fragment, id });
      const data = { ...ev, deleted: true };
      cache.writeFragment({ fragment, id, data });
    },
    optimisticResponse: {
      deleteSilence: {
        deletedId: id,
        __typename: "DeleteRecordPayload",
      },
      __typename: "Mutation",
    },
  });
