import gql from "graphql-tag";

import handle from "/lib/exceptionHandler";

const fragment = gql`
  fragment SetCheckPublishMutation_check on CheckConfig {
    publish
  }
`;

const mutation = gql`
  mutation SetCheckPublishMutation($input: UpdateCheckInput!) {
    updateCheck(input: $input) {
      check {
        ...SetCheckPublishMutation_check
      }
    }
  }

  ${fragment}
`;

export default (client, { id, publish }) =>
  client.mutate({
    mutation,
    variables: {
      input: { id, props: { publish } },
    },
    update: (dataProxy, { data }) => {
      // Apollo wraps the call to `update` in its own try/catch and swallows any
      // errors. We must handle any errors ourselves if we don't want them to
      // be completely silenced.
      try {
        dataProxy.writeFragment({
          id,
          fragment,
          data: data.updateCheck.check,
        });
      } catch (error) {
        handle(error);
      }
    },
    optimisticResponse: {
      updateCheck: {
        check: {
          publish,
          __typename: "CheckConfig",
        },
        __typename: "UpdateCheckPayload",
      },
      __typename: "Mutation",
    },
  });
