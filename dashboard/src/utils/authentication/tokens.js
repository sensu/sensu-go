import moment from "moment";

import createDispatcher from "../dispatcher";

const { subscribe, unsubscribe, subscribeOnce, dispatch } = createDispatcher();

// Instantiate new tokens object with defaults.
export function newTokens(args = {}) {
  return {
    accessToken: args.accessToken,
    refreshToken: args.refreshToken,
    authenticated: args.authenticated || null,
    expiresAt: args.expiresAt || moment(),
  };
}

// Single instance of authentication info.
let authTokens = newTokens();

// Return single authTokens instance
export function get() {
  return authTokens;
}

// Swap single authTokens instance for new one
export function swap(t) {
  if (Object.prototype.toString.call(t) !== "[object Object]" || !t) {
    throw new TypeError("Expected Object");
  }
  authTokens = Object.freeze(t);
  dispatch(authTokens);
}

export { subscribe, unsubscribe, subscribeOnce };
