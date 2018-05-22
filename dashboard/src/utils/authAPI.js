import { parseUNIX } from "/utils/date";
import doFetch from "/utils/fetch";
import { memoize, when } from "/utils/promise";
import { UnauthorizedError } from "/errors/FetchError";

const parseTokenResponse = body => ({
  accessToken: body.access_token,
  refreshToken: body.refresh_token,
  expiresAt: parseUNIX(body.expires_at).toISOString(),
});

export const createTokens = memoize(
  ({ username, password }) => {
    const path = "/auth";
    const config = {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Basic ${window.btoa(`${username}:${password}`)}`,
      },
    };

    return doFetch(path, config)
      .then(response => response.json())
      .then(parseTokenResponse);
  },
  ({ username, password }) => JSON.stringify({ username, password }),
);

export default createTokens;

export const refreshTokens = memoize(
  ({ accessToken, refreshToken }) => {
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

    return doFetch(path, config)
      .then(response => response.json())
      .then(parseTokenResponse);
  },
  ({ accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);

export const invalidateTokens = memoize(
  ({ accessToken, refreshToken }) => {
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
    return doFetch(path, config)
      .catch(when(UnauthorizedError, () => {}))
      .then(() => ({
        accessToken: null,
        refreshToken: null,
        expiresAt: null,
      }));
  },
  ({ accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);
