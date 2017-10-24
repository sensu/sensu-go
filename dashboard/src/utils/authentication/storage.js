import isEmpty from "lodash/isEmpty";
import { newTokens } from "./tokens";

// identifier used to sure tokens under
const authTokensKey = "authTokens";

// persist auth tokens in localStorage
export function persist(args) {
  const tokens = {
    ...args,
    expiresAt: args.expiresAt.unix(),
  };

  // NOTE: Safari does not allow access to localStorage in private browsing mode.
  localStorage.setItem(authTokensKey, JSON.stringify(tokens));
}

// persist auth tokens in localStorage
export function retrieve() {
  const tokens = localStorage.getItem(authTokensKey);
  if (!isEmpty(tokens)) {
    return newTokens(tokens);
  }

  return null;
}
