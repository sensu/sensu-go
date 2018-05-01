// Adds a 'deleted' field to the given type.
function addDeletedFieldTo(typename, { defaultValue = false } = {}) {
  return {
    resolvers: {
      [typename]: {
        deleted: () => defaultValue,
      },
    },
  };
}

export default addDeletedFieldTo;
