import React from "react";
import { makeRouteConfig, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import EventsList from "./components/EventsList";
import ChecksList from "./components/CheckList";

export default makeRouteConfig(
  <Route path="/" Component={AppWrapper}>
    <Route path="events" Component={EventsList} />
    <Route path="checks" Component={ChecksList} />
  </Route>,
);
