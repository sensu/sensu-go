import { parseUNIX } from "/lib/util/date";
import { memoize, when } from "/lib/util/promise";
import { UnauthorizedError } from "/lib/error/FetchError";
import fetch from "/lib/util/fetch";

/*
 * Note on memoization in this file:
 *
 * Each of the auth API request functions is memoized as a mechanism to prevent
 * multiple parallel requests with the same data. If additional API requests are
 * created while an identical one is in-flight, the memoized function will
 * resolve the current pending fetch for all requests.
 *
 * The memoization utility used here is specific for use with functions that
 * return promises - the memoized result for a given key is cleared when the
 * matching promise resolves.
 */

/*
 * Note on keys:
 *
 * Including the value of `cache` in the memoization key would technically
 * be necessary if we weren't able to safely assume that `cache` is a constant
 * value - in our app, where the apollo cache is effectively a global singleton,
 * this is a safe shortcut.
 */

const parseTokenResponse = body => ({
  accessToken: body.access_token,
  refreshToken: body.refresh_token,
  expiresAt: parseUNIX(body.expires_at).toISOString(),
});

export const createTokens = memoize(
  (cache, { username, password }) => {
    const path = "/auth";
    const config = {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Basic ${window.btoa(`${username}:${password}`)}`,
      },
    };

    return fetch(cache)(path, config)
      .then(response => response.json())
      .then(parseTokenResponse);
  },
  // Map arguments to memoization key. See note on keys above.
  (_, { username, password }) => JSON.stringify({ username, password }),
);

export default createTokens;

export const refreshTokens = memoize(
  (cache, { accessToken, refreshToken }) => {
    const path = "/auth/token";
    const config = {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        refresh_token: refreshToken,
      }),
    };

    return fetch(cache)(path, config)
      .then(response => response.json())
      .then(parseTokenResponse);
  },
  // Map arguments to memoization key. See note on keys above.
  (_, { accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);

export const invalidateTokens = memoize(
  (cache, { accessToken, refreshToken }) => {
    const path = "/auth/logout";
    const config = {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        refresh_token: refreshToken,
      }),
    };
    return fetch(cache)(path, config)
      .catch(when(UnauthorizedError, () => {}))
      .then(() => ({
        accessToken: null,
        refreshToken: null,
        expiresAt: null,
      }));
  },
  // Map arguments to memoization key. See note on keys above.
  (_, { accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);
