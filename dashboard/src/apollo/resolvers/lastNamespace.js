export default {
  defaults: {
    lastNamespace: "default",
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
