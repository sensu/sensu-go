import merge from "deepmerge";
import { withClientState } from "apollo-link-state";
import auth from "./resolvers/auth";
import addDeletedFieldTo from "./resolvers/deleted";

const resolvers = merge.all([
  {},
  auth,
  addDeletedFieldTo("Event"),
  addDeletedFieldTo("Entity"),
  addDeletedFieldTo("Silenced"),
]);

const stateLink = ({ cache }) => withClientState({ ...resolvers, cache });
export default stateLink;
