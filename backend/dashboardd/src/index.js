import React from "react";
import ReactDOM from "react-dom";
import { createBrowserRouter } from "found";

import injectTapEventPlugin from "react-tap-event-plugin";
import "typeface-roboto"; // eslint-disable-line import/extensions

import routes from "./routes";
import registerServiceWorker from "./registerServiceWorker";

// Configure Router
const Router = createBrowserRouter({ routeConfig: routes });

// Renderer
ReactDOM.render(<Router />, document.getElementById("root"));

// Register React Tap event plugin
injectTapEventPlugin();

// Register service workers
registerServiceWorker();
