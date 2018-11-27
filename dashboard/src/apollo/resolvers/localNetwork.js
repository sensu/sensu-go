export default {
  defaults: {
    localNetwork: {
      __typename: "LocalNetwork",
      offline: false,
      retry: false,
    },
  },
  resolvers: {
    Mutation: {
      retryLocalNetwork: () => (_, { cache }) => {
        const data = {
          localNetwork: {
            __typename: "LocalNetwork",
            retry: true,
          },
        };
        cache.writeData(data);
        return null;
      },
      setLocalNetworkOffline: (_, { offline }, { cache }) => {
        const data = {
          localNetwork: {
            __typename: "LocalNetwork",
            offline,
            retry: false,
          },
        };
        cache.writeData(data);
        return null;
      },
    },
  },
};
