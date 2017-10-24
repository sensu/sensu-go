import moment from "moment";
import { getAccessToken, authenticate, logout } from "./index";
import * as tokens from "./tokens";
import * as storage from "./storage";

function buildTokens(opts = {}) {
  return tokens.newTokens({
    authenticated: opts.authenticated === undefined || opts.authenticated,
    accessToken: opts.accessToken || "abc",
    expiresAt: opts.expiresAt || moment("2090-10-21 12:00"),
  });
}

describe("getAccessToken", () => {
  beforeEach(() => {
    fetch.resetMocks();
    localStorage.clear();
    tokens.swap(tokens.newTokens({ authenticated: null }));
  });

  it("should retrieve token if already authenticated", async () => {
    const myTokens = buildTokens();
    tokens.swap(myTokens);

    await expect(getAccessToken()).resolves.toBeDefined();
    await expect(getAccessToken()).resolves.toEqual(myTokens.accessToken);
  });

  it("should retrieve token from localStorage if not found", async () => {
    const myTokens = buildTokens();
    storage.persist(myTokens);

    await expect(getAccessToken()).resolves.toBeDefined();
    await expect(getAccessToken()).resolves.toEqual(myTokens.accessToken);
  });

  it("should return null if not authenticated", async () => {
    const myTokens = buildTokens({ authenticated: false });
    tokens.swap(myTokens);

    await expect(getAccessToken()).toBeInstanceOf(Promise);
    await expect(getAccessToken()).resolves.toBeNull();
  });

  it("should return null if not stored in localStorage", async () => {
    const myTokens = buildTokens({ authenticated: null });
    tokens.swap(myTokens);

    await expect(getAccessToken()).toBeInstanceOf(Promise);
    await expect(getAccessToken()).resolves.toBeNull();
  });

  it("should attempt to refresh if token is expireed", async () => {
    const myTokens = buildTokens({ expiresAt: moment("1970-10-01 12:00") });
    tokens.swap(myTokens);

    const newTokens = buildTokens({ accessToken: "12345" });
    fetch.mockResponse(JSON.stringify({ access_token: newTokens.accessToken }));

    await expect(getAccessToken()).toBeInstanceOf(Promise);
    await expect(getAccessToken()).resolves.toBe(newTokens.accessToken);
  });
});

describe("authenticate", () => {
  beforeEach(() => {
    fetch.resetMocks();
    localStorage.clear();
    tokens.swap(tokens.newTokens({ authenticated: null }));
  });

  it("should authenticates user", async () => {
    const newTokens = buildTokens({ accessToken: "12345" });
    fetch.mockResponse(JSON.stringify({ access_token: newTokens.accessToken }));

    await authenticate("test", "pass");
    await expect(getAccessToken()).resolves.toEqual(newTokens.accessToken);
  });

  it("should update state", async () => {
    const newTokens = buildTokens({ accessToken: "12345" });
    fetch.mockResponse(JSON.stringify({ access_token: newTokens.accessToken }));

    await authenticate("test", "pass");
    await expect(tokens.get().accessToken).toEqual(newTokens.accessToken);
    await expect(storage.retrieve().accessToken).toEqual(newTokens.accessToken);
  });

  it("should throw errorr if request fails", async () => {
    fetch.mockResponse(JSON.stringify({}), { status: 401 });

    await expect(authenticate("test", "pass")).rejects.toBeDefined();
    await expect(getAccessToken()).resolves.toBeNull();
  });
});

describe("authenticate", () => {
  beforeEach(() => {
    fetch.resetMocks();
    localStorage.clear();
    tokens.swap(tokens.newTokens({ authenticated: null }));
  });
});
