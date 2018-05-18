import { parseUNIX } from "/utils/date";
import doFetch from "/utils/fetch";
import { memoize } from "/utils/promise";

const parseTokenResponse = body => ({
  accessToken: body.access_token,
  refreshToken: body.refresh_token,
  expiresAt: parseUNIX(body.expires_at).toISOString(),
});

export const createTokens = memoize(
  async ({ username, password }) => {
    const path = "/auth";
    const config = {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Basic ${window.btoa(`${username}:${password}`)}`,
      },
    };

    const response = await doFetch(path, config);
    return parseTokenResponse(await response.json());
  },
  ({ username, password }) => JSON.stringify({ username, password }),
);

export default createTokens;

export const refreshTokens = memoize(
  async ({ accessToken, refreshToken }) => {
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

    const response = await doFetch(path, config);
    return parseTokenResponse(await response.json());
  },
  ({ accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);

export const invalidateTokens = memoize(
  async ({ accessToken, refreshToken }) => {
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

    await doFetch(path, config);

    return {
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
    };
  },
  ({ accessToken, refreshToken }) =>
    JSON.stringify({ accessToken, refreshToken }),
);
