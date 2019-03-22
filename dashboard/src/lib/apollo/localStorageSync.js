import { ApolloError } from "apollo-client";

import handle from "/lib/exceptionHandler";

const getQueryName = document => document.definitions[0].name.value;

const localStorageSync = (client, query, opts = {}) => {
  const { ignoreRemoteUpdates = false } = opts;
  const storageKey = `$SYNC.${getQueryName(query)}`;

  const restoreData = data => {
    try {
      client.writeData({ data: JSON.parse(data) });
    } catch (error) {
      handle(error);
    }
  };

  const currentValue = localStorage.getItem(storageKey);

  if (currentValue) {
    restoreData(currentValue);
  }

  if (!ignoreRemoteUpdates) {
    window.addEventListener("storage", event => {
      if (event.storageArea === localStorage && event.key === storageKey) {
        restoreData(event.newValue);
      }
    });
  }

  const queryObservable = client.watchQuery({ query });
  let first = true;

  queryObservable.subscribe({
    next: () => {
      if (first) {
        first = false;
        return;
      }

      const currentResult = queryObservable.currentResult();

      const { errors, error, data } = currentResult;

      if (error) {
        throw error;
      }

      if (errors && errors.length > 0) {
        throw new ApolloError({ graphQLErrors: errors });
      }

      localStorage.setItem(storageKey, JSON.stringify(data));
    },
    error(error) {
      throw error;
    },
  });
};

export default localStorageSync;
