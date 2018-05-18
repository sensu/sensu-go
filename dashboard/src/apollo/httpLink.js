import { BatchHttpLink as HttpLink } from "apollo-link-batch-http";
import doFetch from "/utils/fetch";

const httpLink = () =>
  new HttpLink({
    uri: "/graphql",
    fetch: doFetch,
    credentials: "same-origin",
    batchMax: 500,
    batchInterval: 5,
  });

export default httpLink;
