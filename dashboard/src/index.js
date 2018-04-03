import React from "react";
import ReactDOM from "react-dom";
import { Provider } from "react-redux";
import { Resolver } from "found-relay";
import FarceActions from "farce/lib/Actions";
import createConnectedRouter from "found/lib/createConnectedRouter";
import createRender from "found/lib/createRender";
import injectTapEventPlugin from "react-tap-event-plugin";

import createStore from "./store";
import reducer from "./reducer";
import registerServiceWorker from "./registerServiceWorker";
import environment from "./environment";
import AppThemeProvider from "./components/AppThemeProvider";
import BrowserOverrides from "./components/BrowserOverrides";
import NotFoundPage from "./pages/NotFoundPage";

// Fonts
import "typeface-roboto"; // eslint-disable-line

// Polyfill
import "url-search-params-polyfill"; // eslint-disable-line import/first

// Configure Router
const Router = createConnectedRouter({
  render: createRender({
    // eslint-disable-next-line react/prop-types
    renderError: ({ error }) => (
      <div>{error.status === 404 ? <NotFoundPage /> : "Error"}</div>
    ),
  }),
});

// Configure store
const store = createStore(reducer, {});
store.dispatch(FarceActions.init());

// Renderer
ReactDOM.render(
  <Provider store={store}>
    <AppThemeProvider>
      <BrowserOverrides />
      <Router resolver={new Resolver(environment)} />
    </AppThemeProvider>
  </Provider>,
  document.getElementById("root"),
);

// Register React Tap event plugin
injectTapEventPlugin();

// Register service workers
registerServiceWorker();
