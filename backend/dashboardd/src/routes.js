import React from "react";
import { makeRouteConfig, Route } from "found";

import App from "./components/App";
import EventsList from "./components/EventsList";
import ChecksList from "./components/CheckList";

export default makeRouteConfig(
  <Route path="/" Component={App}>
    <Route path="events" Component={EventsList} />
    <Route path="checks" Component={ChecksList} />
  </Route>,
);
