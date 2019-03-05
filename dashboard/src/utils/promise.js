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

// Define, but lazily instantiate the memoization cache in order to be sure that
// any necessary polyfills have been installed prior to instantiation.
let cache;

/*
 * Promise-compatible memoization utility
 *
 * This utility is helpful in the situation where the consumers of an async
 * function don't want to be concerned about the expense of the underlying
 * operation (e.g. a network request), and want to be able to call the function
 * indiscriminately whenever a result is needed (possibly continuously).
 *
 * For example, consider a function that sends a fetch request whose result can
 * be considered deterministic for a given set of parameters over the duration
 * of the request (i.e. the result may be different at some time in the future,
 * but over short durations the result will be the same).
 *
 * A standard memoization approach would cache the return value of the request
 * indefinitely, all subsequent requests would return the original response.
 *
 * This approach instead clears the cached return value immediately when the
 * async operation resolves. Subsequent calls will result in new invocations of
 * the wrapped function. This has the effect of serializing a parallel sequence
 * of async requests by combining them.
 */
export const memoize = (promiseCreator, keyCreator) => (...args) => {
  if (!cache) {
    cache = new WeakMap();
  }

  // Calculate the map key for the current function arguments.
  const key = keyCreator(...args);

  // The `cache` WeakMap associates a wrapped function to a map of argument-
  // derived key - returned promise value - pairs.
  let map = cache.get(promiseCreator);

  if (!map) {
    map = new Map();
    cache.set(promiseCreator, map);
  } else if (map.has(key)) {
    // If the map for the current function has a value for the key derived from
    // the current arguments, return that value.
    return map.get(key);
  }

  // No cached value currently exists, call the wrapped function.
  const promise = promiseCreator(...args).then(
    // Delete the cached value for the key derived from the current arguments
    // when the promise resolves (or throws). Return (or throw) the result in
    // to make this operation transparent to the caller.
    result => {
      map.delete(key);
      return result;
    },
    error => {
      map.delete(key);
      throw error;
    },
  );

  // Store the current value for the key derived from the current arguments.
  map.set(key, promise);

  return promise;
};
