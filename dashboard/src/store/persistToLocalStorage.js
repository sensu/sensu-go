import { applyMiddleware } from "redux";
import equals from "ramda/src/equals";

const storage = window.localStorage;

function persist(key, store, prevState) {
  const state = store.getState();
  if (!equals(state[key], prevState[key])) {
    storage.setItem(key, JSON.stringify(state[key]));
  }
}

function middleware(key) {
  let timeoutId = null;
  const clearTimeout = () => {
    window.clearTimeout();
    timeoutId = null;
  };

  return store => next => action => {
    const prevState = store.getState();

    if (typeof timeoutId !== "number") {
      timeoutId = setTimeout(() => {
        persist(key, store, prevState);
        clearTimeout();
      }, 500);
    }
    return next(action);
  };
}

function persistToLocalStorage(key) {
  const enhancer = applyMiddleware(middleware(key));

  return createStore => (...args) => {
    const store = enhancer(createStore)(...args);

    // Retrieve initial state
    const value = JSON.parse(storage.getItem(key) || "{}");
    store.dispatch({ type: "@@storage/CHANGED", payload: value });

    // Set-up listener
    window.addEventListener("storage", e => {
      if (key !== e.key) {
        return;
      }
      store.dispatch({ type: "@@storage/CHANGED", payload: e.newValue });
    });

    return store;
  };
}

export default persistToLocalStorage;
