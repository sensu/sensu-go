export default {
  defaults: {
    lastNamespace: null,
  },
  resolvers: {
    Mutation: {
      setLastNamespace: (_, { name }, { cache }) => {
        cache.writeData({
          data: {
            lastNamespace: name,
          },
        });

        return null;
      },
    },
  },
};
