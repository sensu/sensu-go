// Adds a 'deleted' field to the given type.
//
// https://github.com/apollographql/apollo-client/issues/899
// As best practices for handling deletion with Apollo cache are.. a point of
// discussion, this simple resolver was introudced as a means of allowing for
// soft-deletion on the client side.
//
// NOTE: Currently if a record was deleted and was later re-created by another
// user and refetched the client, value of the deleted field will persist. As
// such in the long run this resolver would be bolstered by the addition of a
// link or middleware that can reset the field when a new instance is received
// from the server (or whatever the source of truth may be.)
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
