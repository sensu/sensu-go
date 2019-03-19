import merge from "deepmerge";
import { withClientState } from "apollo-link-state";
import auth from "./resolvers/auth";
import lastNamespace from "./resolvers/lastNamespace";
import localNetwork from "./resolvers/localNetwork";
import addDeletedFieldTo from "./resolvers/deleted";

const resolvers = merge.all([
  {},
  auth,
  addDeletedFieldTo("CheckConfig"),
  addDeletedFieldTo("Entity"),
  addDeletedFieldTo("Event"),
  addDeletedFieldTo("Silenced"),
  addDeletedFieldTo("CheckConfig"),
  lastNamespace,
  localNetwork,
]);

const stateLink = ({ cache }) => withClientState({ ...resolvers, cache });
export default stateLink;
