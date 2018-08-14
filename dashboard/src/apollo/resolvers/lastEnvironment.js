export default {
  defaults: {
    lastEnvironment: null,
  },
  resolvers: {
    Mutation: {
      setLastEnvironment: (_, { environment, organization }, { cache }) => {
        cache.writeData({
          data: {
            lastEnvironment: {
              __typename: "Namespace",
              environment,
              organization,
            },
          },
        });

        return null;
      },
    },
  },
};
