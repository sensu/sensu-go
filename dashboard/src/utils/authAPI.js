import { parseUNIX } from "./date";

const doFetch = async (path, config) => {
  const response = await fetch(path, config);

  if (response.status < 200 || response.status >= 300) {
    const error = new Error(response.status);
    error.response = response;
    throw error;
  }

  return response;
};

const parseTokenResponse = body => ({
  accessToken: body.access_token,
  refreshToken: body.refresh_token,
  expiresAt: parseUNIX(body.expires_at).toISOString(),
});

export const createTokens = async ({ username, password }) => {
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
};

export default createTokens;

export const refreshTokens = async ({ accessToken, refreshToken }) => {
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
};

export const invalidateTokens = async ({ accessToken, refreshToken }) => {
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
};
