import gql from "graphql-tag";
import { isApolloError } from "apollo-client/errors/ApolloError";

const RESOURCE_EXISTS_ERROR =
  "GraphQL error: error: code = 3 desc = resource already exists";

const fragment = gql`
  fragment CreateSilenceMutation_silence on Silenced {
    id
    name
    expire
    expires
    expireOnResolve
    reason
    subscription
    begin
  }
`;

const mutation = gql`
  mutation CreateSilenceMutation($input: CreateSilenceInput!) {
    createSilence(input: $input) {
      silence {
        ...CreateSilenceMutation_silence
      }
    }
  }

  ${fragment}
`;

export default (client, input) =>
  client
    .mutate({
      mutation,
      variables: {
        input,
      },
    })
    .catch(error => {
      // HACK: Transform query root-level "resource already exists" error to
      // mutation specific validation errors. This temporarily simulates how we
      // intend to report this error state in the future and can be removed once
      // the API has been updated.
      if (isApolloError(error) && error.message === RESOURCE_EXISTS_ERROR) {
        return {
          data: {
            createSilence: {
              silence: null,
              errors: [
                {
                  code: "VALIDATION_UNIQUE_CONSTRAINT",
                  input: "check",
                },
                {
                  code: "VALIDATION_UNIQUE_CONSTRAINT",
                  input: "subscription",
                },
              ],
            },
          },
        };
      }

      throw error.networkError || error;
    })
    .then(result => result.data.createSilence);
