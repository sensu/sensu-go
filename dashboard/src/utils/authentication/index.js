import defer from "lodash/defer";
import identity from "lodash/fp/identity";
import moment from "moment";

import { requestNewTokens, refreshTokens, invalidateTokens } from "./requests";
import * as tokens from "./tokens";
import * as storage from "./storage";

// Swap global instance and update localStorage
function updateState(newTokens) {
  storage.persist(newTokens);
  defer(() => tokens.swap(newTokens));
}

// Returns a promise that resolves to instance's access token; transparently
// handles refreshing access token if required.
export function getAccessToken() {
  const now = moment();
  const authTokens = tokens.get();

  if (authTokens.authenticated) {
    // Return access token if it is present and has not expired
    if (now.isAfter(authTokens.expiresAt)) {
      return Promise.resolve(authTokens.accessToken);
    }

    // When expired, attempt to refresh the token and return result
    const refresh = refreshTokens(authTokens);
    return refresh.then(newTokens => newTokens.accessToken);
  }

  // If the status is null then attempt to pull token from localStorage
  if (authTokens.authenticated === null) {
    const storedTokens = storage.retrieve();
    if (storedTokens) {
      tokens.swap(storage.retrieve());
      return getAccessToken();
    }
  }

  return Promise.resolve(null);
}

// Sends authentication request to backend and then updates state & storage.
export function authenticate(username, password) {
  // No-op when instance is already authenticated
  const authTokens = tokens.get();
  if (authTokens.authenticated) {
    return Promise.resolve({});
  }

  // Request new auth tokens from the backend and update state
  const requestPromise = requestNewTokens(username, password);
  return requestPromise.then(identity(updateState));
}

// Logout clears state, storage and sends invalidation requets to backend.
export function logout() {
  const authTokens = tokens.get();
  const invalidatePromise = invalidateTokens(authTokens);

  const newTokens = tokens.newTokens({ authenticated: false });
  tokens.swap(newTokens);
  storage.persist(newTokens);

  return invalidatePromise;
}
