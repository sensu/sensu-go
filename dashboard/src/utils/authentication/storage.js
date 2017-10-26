import moment from "moment";
import isEmpty from "lodash/isEmpty";
import { newTokens } from "./tokens";

// identifier used to sure tokens under
const authTokensKey = "authTokens";

// persist auth tokens in localStorage
export function persist(args) {
  const tokens = {
    ...args,
    expiresAt: args.expiresAt.toJSON(),
  };

  // NOTE: Safari does not allow access to localStorage in private browsing mode.
  localStorage.setItem(authTokensKey, JSON.stringify(tokens));
}

// persist auth tokens in localStorage
export function retrieve() {
  const json = localStorage.getItem(authTokensKey);
  if (!isEmpty(json)) {
    const tokens = JSON.parse(json);
    return newTokens({
      ...tokens,
      expiresAt: moment(tokens.expiresAt, moment.ISO_8601),
    });
  }

  return null;
}

export function clear() {
  localStorage.removeItem(authTokensKey);
}
