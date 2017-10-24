import React from "react";
import { makeRouteConfig, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import LoginPage from "./components/Login";
import EventsList from "./components/EventsList";
import ChecksList from "./components/CheckList";

export default makeRouteConfig(
  <Route>
    <Route path="/login" Component={LoginPage} />
    <Route path="/" Component={AppWrapper}>
      <Route path="events" Component={EventsList} />
      <Route path="checks" Component={ChecksList} />
    </Route>
  </Route>,
);
