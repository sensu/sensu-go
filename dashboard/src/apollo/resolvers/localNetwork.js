// @flow
import type { ApolloCache } from "react-apollo";

type Context = { cache: ApolloCache<mixed> };

export const setOffline = (cache: ApolloCache<mixed>, offline: boolean) => {
  const data = {
    localNetwork: {
      __typename: "LocalNetwork",
      offline,
      retry: false,
    },
  };

  cache.writeData({ data });
};

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
      retryLocalNetwork: (_: mixed, args: mixed, { cache }: Context) => {
        const data = {
          localNetwork: {
            __typename: "LocalNetwork",
            retry: true,
          },
        };
        cache.writeData({ data });
        return null;
      },
      setLocalNetworkOffline: (
        _: mixed,
        { offline }: { offline: boolean },
        { cache }: Context,
      ) => {
        setOffline(cache, offline);
        return null;
      },
    },
  },
};
