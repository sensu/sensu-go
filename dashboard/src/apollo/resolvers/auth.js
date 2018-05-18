import gql from "graphql-tag";

import { createTokens, invalidateTokens, refreshTokens } from "/utils/authAPI";

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
      createTokens: (_, { username, password }, { cache }) =>
        createTokens({ username, password }).then(tokens => {
          const { accessToken, refreshToken, expiresAt } = tokens;
          const data = {
            __typename: "CreateTokensMutation",
            auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
          };

          cache.writeData({ data });

          return data;
        }),

      refreshTokens: (_, { notBefore = null }, { cache }) => {
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

        return refreshTokens(result.auth).then(tokens => {
          const { accessToken, refreshToken, expiresAt } = tokens;
          const data = {
            __typename: "RefreshTokensMutation",
            auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
          };

          cache.writeData({ data });

          return data;
        });
      },

      invalidateTokens: (_, vars, { cache }) => {
        const result = cache.readQuery({ query });
        return invalidateTokens(result.auth).then(tokens => {
          const { accessToken, refreshToken, expiresAt } = tokens;
          const data = {
            __typename: "InvalidateTokensMutation",
            auth: { __typename: "Auth", accessToken, refreshToken, expiresAt },
          };

          cache.writeData({ data });

          return data;
        });
      },
    },
  },
};
