import { createStore as create, applyMiddleware } from "redux";
import { compose } from "lodash/fp";
import { devToolsEnhancer } from "redux-devtools-extension/logOnlyInProduction";

import thunkMiddleware from "redux-thunk";
import persistToLocalStorage from "./persistToLocalStorage";

function createStore(reducer) {
  const enhancer = compose(
    applyMiddleware(
      thunkMiddleware,
      ...(process.env.NODE_ENV === "production"
        ? // eslint-disable-next-line global-require
          [require("redux-freeze")]
        : []),
    ),
    persistToLocalStorage("theme"),
    devToolsEnhancer({ title: "Sensu Web UI" }),
  );

  return create(reducer, enhancer);
}

export default createStore;
