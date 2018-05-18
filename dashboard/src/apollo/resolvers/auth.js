import gql from "graphql-tag";

import { when } from "/utils/promise";
import { createTokens, invalidateTokens, refreshTokens } from "/utils/authAPI";

import { UnauthorizedError } from "/errors/FetchError";

const query = gql`
  query AuthQuery {
    auth @client {
      accessToken
      refreshToken
      expiresAt
    }
  }
`;

export default {
  defaults: {
    auth: {
      __typename: "Auth",
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
    },
  },
  resolvers: {
    Mutation: {
      createTokens: async (_, { username, password }, { cache }) => {
        const tokens = await createTokens({ username, password });

        const { accessToken, refreshToken, expiresAt } = tokens;
        const data = {
          __typename: "CreateTokensMutation",
          auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
        };

        cache.writeData({ data });

        return data;
      },

      refreshTokens: async (_, { notBefore = null }, { cache }) => {
        const result = cache.readQuery({ query });

        if (notBefore !== null && isNaN(new Date(notBefore))) {
          throw new TypeError(
            "invalid `notBefore` variable. Expected DateTime",
          );
        }

        const expired =
          !notBefore ||
          !result.auth.expiresAt ||
          new Date(notBefore) > new Date(result.auth.expiresAt);

        if (!expired) {
          return {
            __typename: "RefreshTokensMutation",
            ...result,
          };
        }

        const tokens = await refreshTokens(result.auth);

        const { accessToken, refreshToken, expiresAt } = tokens;
        const data = {
          __typename: "RefreshTokensMutation",
          auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
        };

        cache.writeData({ data });

        return data;
      },

      invalidateTokens: async (_, vars, { cache }) => {
        const result = cache.readQuery({ query });
        const tokens = await invalidateTokens(result.auth).catch(
          when(UnauthorizedError, () => ({
            accessToken: null,
            refreshToken: null,
            expiresAt: null,
          })),
        );

        const { accessToken, refreshToken, expiresAt } = tokens;
        const data = {
          __typename: "InvalidateTokensMutation",
          auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
        };

        cache.writeData({ data });

        return data;
      },
    },
  },
};
