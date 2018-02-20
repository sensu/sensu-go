import { createStore as create, applyMiddleware } from "redux";
import { compose, identity } from "lodash/fp";
import BrowserProtocol from "farce/lib/BrowserProtocol";
import createHistoryEnhancer from "farce/lib/createHistoryEnhancer";
import queryMiddleware from "farce/lib/queryMiddleware";
import createMatchEnhancer from "found/lib/createMatchEnhancer";
import Matcher from "found/lib/Matcher";

import thunkMiddleware from "redux-thunk";
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
    createHistoryEnhancer({
      protocol: new BrowserProtocol(),
      middlewares: [queryMiddleware],
    }),
    createMatchEnhancer(new Matcher(routeConfig)),
    identity,
  );

  return create(reducer, enhancer);
}

export default createStore;
