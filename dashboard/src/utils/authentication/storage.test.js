import moment from "moment";
import * as storage from "./storage";
import { newTokens } from "./tokens";

describe("persist", () => {
  afterEach(() => localStorage.clear());

  it("should store given tokens in localStorage", () => {
    const myTokens = newTokens({ accessToken: "abc", expiresAt: moment() });
    expect(storage.retrieve()).toBeNull();
    storage.persist(myTokens);

    const stored = storage.retrieve();
    expect(stored.accessToken).toEqual(myTokens.accessToken);
    expect(stored.expiresAt.toJSON()).toEqual(myTokens.expiresAt.toJSON());
  });
});

describe("retrieve", () => {
  afterEach(() => localStorage.clear());
  it("should return null if no tokens stored", () => {
    expect(storage.retrieve()).toBeNull();
  });
  it("should retreieve stored tokens and wrap date", () => {
    const myTokens = newTokens({ accessToken: "abc", expiresAt: moment() });
    storage.persist(myTokens);

    const stored = storage.retrieve();
    expect(stored.accessToken).toEqual(myTokens.accessToken);
    expect(stored.refreshToken).toEqual(myTokens.refreshToken);
    expect(stored.authenticated).toEqual(myTokens.authenticated);
    expect(stored.expiresAt.toJSON()).toEqual(myTokens.expiresAt.toJSON());
    expect(stored.expiresAt).toBeInstanceOf(moment);
  });
});
