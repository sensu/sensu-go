import { BatchHttpLink as HttpLink } from "apollo-link-batch-http";
import fetch from "/lib/util/fetch";

const httpLink = ({ getClient }) =>
  new HttpLink({
    uri: "/graphql",
    // We need to defer the call to `getClient().cache` until after the Apollo
    // client is initialized - hence this meaty wrapper around `fetch`.
    fetch: (...args) => fetch(getClient().cache)(...args),
    credentials: "same-origin",
    batchMax: 10,
    batchInterval: 3,
  });

export default httpLink;
