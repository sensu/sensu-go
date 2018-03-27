import { createStore as create, applyMiddleware } from "redux";
import { compose } from "lodash/fp";
import { devToolsEnhancer } from "redux-devtools-extension/logOnlyInProduction";
import BrowserProtocol from "farce/lib/BrowserProtocol";
import createHistoryEnhancer from "farce/lib/createHistoryEnhancer";
import queryMiddleware from "farce/lib/queryMiddleware";
import createMatchEnhancer from "found/lib/createMatchEnhancer";
import Matcher from "found/lib/Matcher";

import thunkMiddleware from "redux-thunk";
import persistToLocalStorage from "./persistToLocalStorage";
import routeConfig from "../routes";

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
    createHistoryEnhancer({
      protocol: new BrowserProtocol(),
      middlewares: [queryMiddleware],
    }),
    createMatchEnhancer(new Matcher(routeConfig)),
    devToolsEnhancer({ title: "Sensu Web UI" }),
  );

  return create(reducer, enhancer);
}

export default createStore;
