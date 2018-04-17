import createDispatcher from "./dispatcher";

describe("dispatcher", () => {
  describe("dispatch", () => {
    it("should notify listeners", () => {
      const dispatcher = createDispatcher();

      const listener1 = jest.fn();
      const listener2 = jest.fn();

      dispatcher.subscribe(listener1);
      dispatcher.subscribe(listener2);

      const data = { foo: "bar" };
      dispatcher.dispatch(data);

      expect(listener1.mock.calls.length).toBe(1);
      expect(listener2.mock.calls.length).toBe(1);

      expect(listener1.mock.calls[0][0]).toBe(data);
      expect(listener2.mock.calls[0][0]).toBe(data);

      dispatcher.dispatch(data);

      expect(listener1.mock.calls.length).toBe(2);
      expect(listener2.mock.calls.length).toBe(2);
    });
  });

  describe("subscribe", () => {
    it("should be idempotent", () => {
      const dispatcher = createDispatcher();

      const listener = jest.fn();

      dispatcher.subscribe(listener);
      dispatcher.subscribe(listener);
      dispatcher.subscribe(listener);

      const data = { foo: "bar" };
      dispatcher.dispatch(data);

      expect(listener.mock.calls.length).toBe(1);
    });
  });

  describe("unsubscribe", () => {
    it("should remove listeners", () => {
      const dispatcher = createDispatcher();

      const listener1 = jest.fn();
      const listener2 = jest.fn();

      const unsubscribe1 = dispatcher.subscribe(listener1);
      dispatcher.subscribe(listener2);

      const data = { foo: "bar" };
      dispatcher.dispatch(data);

      expect(listener1.mock.calls.length).toBe(1);
      expect(listener2.mock.calls.length).toBe(1);

      unsubscribe1();
      dispatcher.unsubscribe(listener2);

      dispatcher.dispatch(data);

      expect(listener1.mock.calls.length).toBe(1);
      expect(listener2.mock.calls.length).toBe(1);
    });

    it("should be idempotent", () => {
      const dispatcher = createDispatcher();

      const listener = jest.fn();

      dispatcher.subscribe(listener);

      const data = { foo: "bar" };
      dispatcher.dispatch(data);

      expect(listener.mock.calls.length).toBe(1);

      dispatcher.unsubscribe(listener);
      dispatcher.unsubscribe(listener);
      dispatcher.unsubscribe(listener);

      dispatcher.dispatch(data);

      expect(listener.mock.calls.length).toBe(1);
    });
  });

  describe("subscribeOnce", () => {
    it("should remove listeners after dispatch", () => {
      const dispatcher = createDispatcher();

      const listener = jest.fn();

      dispatcher.subscribeOnce(listener);

      const data = { foo: "bar" };
      dispatcher.dispatch(data);

      expect(listener.mock.calls.length).toBe(1);
      dispatcher.dispatch(data);

      expect(listener.mock.calls.length).toBe(1);
    });
  });
});
