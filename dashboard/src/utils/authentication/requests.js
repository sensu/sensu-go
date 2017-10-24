import moment from "moment";
import { newTokens } from "./tokens";

const authPath = "/auth";
const refreshPath = "/auth/tokens";
const invalidatePath = "/auth/logout";

function checkStatus(response) {
  if (response.status >= 200 && response.status < 300) {
    return response;
  }

  const err = new Error(response.status);
  err.response = response;
  return Promise.reject(response);
}

function newTokensFromJSON(json) {
  return newTokens({
    accessToken: json.access_token,
    refreshToken: json.refresh_token,
    expiresAt: moment(json.expires_at, "X"),
    authenticated: true,
  });
}

// Request new authentication tokens from backend
export function requestNewTokens(username, password) {
  const authInfo = window.btoa(`${username}:${password}`);
  const fetchPromise = fetch(authPath, {
    method: "GET",
    headers: {
      Accept: "application/json",
      Authorization: `Basic ${authInfo}`,
    },
  });

  return fetchPromise
    .then(checkStatus)
    .then(res => res.json())
    .then(newTokensFromJSON);
}

// Using refresh token request new authentication tokens from backend
export function refreshTokens(tokens) {
  const fetchPromise = fetch(refreshPath, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      Authorization: `Bearer ${tokens.accessToken}`,
    },
    body: JSON.stringify({
      refresh_token: tokens.refreshToken,
    }),
  });

  return fetchPromise
    .then(checkStatus)
    .then(res => res.json())
    .then(newTokensFromJSON);
}

// Using refresh token request new authentication tokens from backend
export function invalidateTokens(tokens) {
  const fetchPromise = fetch(invalidatePath, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      Authorization: `Bearer ${tokens.accessToken}`,
    },
    body: JSON.stringify({
      refresh_token: tokens.refreshToken,
    }),
  });

  return fetchPromise.then(checkStatus);
}
