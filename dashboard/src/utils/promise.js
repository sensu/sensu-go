// When is a utility to allow catching specific types of errors in a promise
// callback. Errors not matching the defined types are re-thrown to be handled
// later in the chain.
//
// When takes an unlimited number of type/hander argument pairs. A type
// argument can be either an error class or an array of error classes, when an
// error matching the type or array of types is caught, the matching handler is
// called.
export const when = (...args) => {
  const handlers = args.reduce((result, arg) => {
    const current = result[result.length - 1];
    if (current && current.length === 1) {
      current.push(arg);
    } else if (Array.isArray(arg)) {
      result.push([arg]);
    } else {
      result.push([[arg]]);
    }
    return result;
  }, []);

  return error => {
    for (let i = 0; i < handlers.length; i += 1) {
      const types = handlers[i][0];
      for (let j = 0; j < types.length; j += 1) {
        if (error instanceof types[j]) {
          return handlers[i][1](error);
        }
      }
    }

    throw error;
  };
};

// TODO: Ensure WeakMap is polyfilled
const cache = new WeakMap();

export const memoize = (promiseCreator, keyCreator) => (...args) => {
  let map = cache.get(promiseCreator);
  const key = keyCreator(...args);

  if (!map) {
    map = new Map();
    cache.set(promiseCreator, map);
  } else if (map.has(key)) {
    return map.get(key);
  }

  const promise = promiseCreator(...args).then(
    result => {
      map.delete(key);
      return result;
    },
    error => {
      map.delete(key);
      throw error;
    },
  );

  map.set(key, promise);

  return promise;
};
