import { setContext } from "apollo-link-context";

import handle from "/exceptionHandler";

import refreshTokens from "/mutations/refreshTokens";

const EXPIRY_THRESHOLD_MS = 13 * 60 * 1000;

const authLink = ({ getClient }) =>
  setContext(async () => {
    try {
      const notBefore = new Date(Date.now() + EXPIRY_THRESHOLD_MS);

      // TODO: Prevent parallel auth token requests
      const { data } = await refreshTokens(getClient(), {
        notBefore: notBefore.toISOString(),
      });

      return {
        headers: {
          Authorization: `Bearer ${data.refreshTokens.auth.accessToken}`,
        },
      };
    } catch (error) {
      handle(error);
      return {};
    }
  });

export default authLink;
