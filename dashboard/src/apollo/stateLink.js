import { withClientState } from "apollo-link-state";
import merge from "lodash/merge";

import auth from "./resolvers/auth";

const resolvers = merge(auth);

const stateLink = ({ cache }) => withClientState({ ...resolvers, cache });

export default stateLink;
