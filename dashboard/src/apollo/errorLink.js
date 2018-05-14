import { onError } from "apollo-link-error";

import handle from "/exceptionHandler";

const errorLink = () =>
  onError(({ graphQLErrors, networkError }) => {
    // TODO: Connect this error handler to display a blocking error alert
    if (graphQLErrors)
      graphQLErrors.forEach(error => {
        handle(error.originalError || error);
      });
    if (networkError) handle(networkError.networkError || networkError);
  });

export default errorLink;
