import { Environment, Network, RecordSource, Store } from "relay-runtime";
import { getAccessToken } from "./utils/authentication";

function fetchQuery(
  operation,
  variables,
  // cacheConfig,
  // uploadables,
) {
  const accessToken = getAccessToken();
  const request = fetch("/graphql", {
    method: "POST",
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${accessToken}`,
      "content-type": "application/json",
    },
    body: JSON.stringify({
      query: operation.text,
      variables,
    }),
  });

  return request.then(res => res.json());
}

// Create a record source & instantiate store
const source = new RecordSource();
const store = new Store(source);

// Create a network layer from the fetch function
const network = Network.create(fetchQuery);

// Create an environment using this network:
const environment = new Environment({
  network,
  source,
  store,
});

export default environment;
