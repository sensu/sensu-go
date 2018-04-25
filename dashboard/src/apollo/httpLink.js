import { HttpLink } from "apollo-link-http";

const httpLink = () =>
  new HttpLink({
    uri: "/graphql",
    fetchOptions: {},
    credentials: "same-origin",
  });

export default httpLink;
