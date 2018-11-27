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
      retryLocalNetwork: () => null,
      setLocalNetworkOffline: () => null,
    },
  },
};
