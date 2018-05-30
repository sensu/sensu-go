const createDispatcher = () => {
  const listeners = [];

  const dispatch = data => {
    [...listeners].forEach(listener => listener(data));
  };

  const unsubscribe = listener => {
    const index = listeners.indexOf(listener);
    if (index !== -1) {
      listeners.splice(index, 1);
    }
  };

  const subscribe = listener => {
    if (listeners.indexOf(listener) !== -1) {
      return () => {};
    }
    listeners.push(listener);

    return () => unsubscribe(listener);
  };

  const subscribeOnce = handler => {
    let unsub;

    const listener = data => {
      handler(data);
      unsub();
    };

    unsub = subscribe(listener);
    return unsub;
  };

  return {
    dispatch,
    unsubscribe,
    subscribe,
    subscribeOnce,
  };
};

export default createDispatcher;
