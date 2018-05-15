import { HttpLink } from "apollo-link-http";

import doFetch from "/utils/fetch";

const httpLink = () =>
  new HttpLink({
    uri: "/graphql",
    fetch: doFetch,
    credentials: "same-origin",
  });

export default httpLink;
