import { Environment, RecordSource, Store } from "relay-runtime";
import {
  RelayNetworkLayer,
  loggerMiddleware,
  errorMiddleware,
  perfMiddleware,
  retryMiddleware,
  authMiddleware,
} from "react-relay-network-modern";
import { getAccessToken } from "./utils/authentication";

// Create a record source & instantiate store
const source = new RecordSource();
const store = new Store(source);

// Create a network layer from the fetch function
const network = new RelayNetworkLayer([
  ...(process.env.NODE_ENV !== "production"
    ? [loggerMiddleware(), errorMiddleware(), perfMiddleware()]
    : []),
  authMiddleware({
    token: getAccessToken,
  }),
  retryMiddleware({
    statusCodes: code => code >= 200 && code < 400,
  }),
]);

// Create an environment using this network:
const environment = new Environment({
  network,
  source,
  store,
});

export default environment;
