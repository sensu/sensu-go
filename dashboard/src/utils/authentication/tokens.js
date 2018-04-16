import moment from "moment";

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
  authTokens = t;
}

// Initialize authTokens w/ empty instance
swap(newTokens());
