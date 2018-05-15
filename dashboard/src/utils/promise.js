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
    for (let i = 0; i < handlers.length; i++) {
      const types = handlers[i][0];
      for (let j = 0; j < types.length; j++) {
        if (error instanceof types[j]) {
          return handlers[i][1](error);
        }
      }
    }

    throw error;
  };
};
