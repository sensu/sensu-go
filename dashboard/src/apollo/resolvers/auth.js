import gql from "graphql-tag";

import { createTokens, invalidateTokens, refreshTokens } from "/utils/authAPI";
import { when } from "/utils/promise";
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

const handleTokens = (cache, typename) => tokens => {
  const { accessToken, refreshToken, expiresAt } = tokens;
  const data = {
    __typename: typename,
    auth: {
      __typename: "Auth",
      invalid: false,
      accessToken,
      refreshToken,
      expiresAt,
    },
  };

  cache.writeData({ data });

  return data;
};

const handleError = (cache, typename) =>
  when(UnauthorizedError, error => {
    const data = {
      __typename: typename,
      auth: {
        __typename: "Auth",
        invalid: true,
      },
    };

    cache.writeData({ data });

    throw error;
  });

export default {
  defaults: {
    auth: {
      __typename: "Auth",
      invalid: false,
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
    },
  },
  resolvers: {
    Mutation: {
      createTokens: (_, { username, password }, { cache }) =>
        createTokens(cache, {
          username,
          password,
        }).then(
          handleTokens(cache, "CreateTokensMutation"),
          handleError(cache, "CreateTokensMutation"),
        ),

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

        return refreshTokens(cache, result.auth).then(
          handleTokens(cache, "RefreshTokensMutation"),
          handleError(cache, "RefreshTokensMutation"),
        );
      },

      invalidateTokens: (_, vars, { cache }) => {
        const result = cache.readQuery({ query });

        // Reset all data in the client cache.
        cache.reset();

        return invalidateTokens(cache, result.auth).then(
          handleTokens(cache, "InvalidateTokensMutation"),
        );
      },

      flagTokens: (_, vars, { cache }) => {
        const data = {
          auth: {
            __typename: "Auth",
            invalid: true,
          },
        };
        cache.writeData({ data });

        return null;
      },
    },
  },
};
