import { withClientState } from "apollo-link-state";
import auth from "./resolvers/auth";
import addDeletedFieldTo from "./resolvers/deleted";

function mergeResolvers(...resolvers) {
  return resolvers.reduce(
    (acc, res) => ({
      defaults: Object.assign({}, acc.defaults, res.defaults),
      resolvers: Object.assign({}, acc.resolvers, res.resolvers),
    }),
    {},
  );
}

const resolvers = mergeResolvers(
  auth,
  addDeletedFieldTo("Event"),
  addDeletedFieldTo("Entity"),
);

const stateLink = ({ cache }) => withClientState({ ...resolvers, cache });
export default stateLink;
