import { when, memoize } from "./promise";

describe("promise utils", () => {
  describe("when", () => {
    it("should throw received error", () => {
      expect(() => {
        when()(new Error());
      }).toThrow(Error);
    });

    it("should call first matching handler for error type", () => {
      const handler = jest.fn();
      const otherHandler = jest.fn();
      const error = new Error();
      when(TypeError, otherHandler, Error, handler, Error, otherHandler)(error);

      expect(handler.mock.calls).toHaveLength(1);
      expect(handler.mock.calls[0][0]).toBe(error);
    });

    it("should support arrays of types", () => {
      const handler = jest.fn();
      const otherHandler = jest.fn();
      const error = new Error();

      when(
        TypeError,
        otherHandler,
        [TypeError, Error],
        handler,
        Error,
        otherHandler,
      )(error);

      expect(handler.mock.calls).toHaveLength(1);
      expect(handler.mock.calls[0][0]).toBe(error);
    });
  });

  describe("memoize", () => {
    let creator;
    let key;

    beforeEach(() => {
      creator = jest.fn(val => Promise.resolve(val));
      key = jest.fn(val => val);
    });

    it("should call the wrapped functions with provided args ", async () => {
      await memoize(creator, key)("arg1", "arg2");

      expect(creator.mock.calls).toHaveLength(1);
      expect(creator.mock.calls[0]).toEqual(["arg1", "arg2"]);

      expect(key.mock.calls).toHaveLength(1);
      expect(key.mock.calls[0]).toEqual(["arg1", "arg2"]);
    });

    it("should return the result of the promise creator", async () => {
      const data = {};
      expect(await memoize(creator, key)(data)).toBe(data);
    });

    it("should return the cached promise instance for a matching key", async () => {
      const wrapped = memoize(creator, key);

      wrapped("a");
      await wrapped("a");

      expect(creator.mock.calls).toHaveLength(1);
    });

    it("should return a new promise without a cache key match", async () => {
      const wrapped = memoize(creator, key);

      wrapped("a");
      await wrapped("b");

      expect(creator.mock.calls).toHaveLength(2);
    });

    it("should clear cache for key when promise resolves", async () => {
      const wrapped = memoize(creator, key);

      await wrapped("a");
      await wrapped("a");

      expect(creator.mock.calls).toHaveLength(2);
    });

    it("should clear cache for key when promise rejects", async () => {
      creator = jest.fn(val => Promise.reject(val));
      const wrapped = memoize(creator, key);

      await wrapped("a").catch(() => {});
      await wrapped("a").catch(() => {});

      expect(creator.mock.calls).toHaveLength(2);
    });
  });
});
