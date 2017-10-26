import * as tokens from "./tokens";

describe("get", () => {
  it("should return singleton instance", () => {
    const result = tokens.get();
    expect(result).toBeTruthy();
  });
});

describe("swap", () => {
  it("should swap singleton instance", () => {
    const myTokens = { testing: 123 };
    expect(tokens.get()).not.toBe(myTokens);
    tokens.swap(myTokens);
    expect(tokens.get()).toBe(myTokens);
  });
});

describe("newTokens", () => {
  it("should return tokens object w/ defaults given no arguments", () => {
    const myTokens = tokens.newTokens();
    expect(myTokens.expiresAt).not.toBeUndefined();
    expect(myTokens.authenticated).not.toBeUndefined();
    expect(myTokens.authenticated).toBeNull();
  });

  it("should return tokens object when arguments", () => {
    const myTokens = tokens.newTokens({
      authenticated: true,
      accessToken: "abc",
    });
    expect(myTokens.authenticated).toBe(true);
    expect(myTokens.accessToken).toBe("abc");
    expect(myTokens.expiresAt).toBeTruthy();
  });
});
