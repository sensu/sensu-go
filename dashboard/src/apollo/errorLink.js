import { onError } from "apollo-link-error";

const errorLink = () =>
  onError(({ graphQLErrors, networkError }) => {
    // TODO: Connect this error handler to display a blocking error alert
    if (graphQLErrors)
      graphQLErrors.forEach(error => {
        // eslint-disable-next-line no-console
        console.error(error.originalError || error);
      });
    // eslint-disable-next-line no-console
    if (networkError) console.error(networkError);
  });

export default errorLink;
