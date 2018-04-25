import { setContext } from "apollo-link-context";
import refreshTokens from "/mutations/refreshTokens";

const EXPIRY_THRESHOLD_MS = 13 * 60 * 1000;

const authLink = ({ getClient }) =>
  setContext(async () => {
    const notBefore = new Date(Date.now() + EXPIRY_THRESHOLD_MS);

    const { data } = await refreshTokens(getClient(), {
      notBefore: notBefore.toISOString(),
    });

    return {
      headers: {
        Authorization: `Bearer ${data.refreshTokens.auth.accessToken}`,
      },
    };
  });

export default authLink;
