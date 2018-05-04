import { withClientState } from "apollo-link-state";
import merge from "lodash/merge";

import auth from "./resolvers/auth";
import addDeletedFieldTo from "./resolvers/deleted";

const resolvers = merge(auth, addDeletedFieldTo("Event"));

const stateLink = ({ cache }) => withClientState({ ...resolvers, cache });

export default stateLink;
