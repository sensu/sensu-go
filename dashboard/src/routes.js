import React from "react";
import { makeRouteConfig, Redirect, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import LoginPage from "./pages/Login";
import EventsPage from "./pages/EventsPage";
import ChecksPage from "./pages/ChecksPage";

export default makeRouteConfig(
  <Route>
    <Route path="/login" Component={LoginPage} />
    <Route
      path="/:organization/:environment"
      Component={AppWrapper}
      query={AppWrapper.query}
    >
      <Route path="events" Component={EventsPage} query={EventsPage.query} />
      <Route path="checks" Component={ChecksPage} query={ChecksPage.query} />
      <Redirect from="dashboard" to="" />
    </Route>
  </Route>,
);
