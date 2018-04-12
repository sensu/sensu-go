import React from "react";
import { makeRouteConfig, Redirect, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import RestrictUnauthenticated from "./components/RestrictUnauthenticated";

import ChecksPage from "./pages/ChecksPage";
import DashboardPage from "./pages/DashboardPage";
import EventsPage from "./pages/EventsPage";
import LoginPage from "./pages/Login";
import RootRedirect from "./pages/RootRedirect";

export default makeRouteConfig(
  <Route>
    <Route path="/login" Component={LoginPage} />
    <Route path="/" Component={RootRedirect} />

    <Route Component={RestrictUnauthenticated}>
      <Route path="/:organization/:environment" Component={AppWrapper}>
        <Route path="" Component={DashboardPage} />
        <Route path="checks" Component={ChecksPage} />
        <Route path="events" Component={EventsPage} query={EventsPage.query} />
        <Redirect from="dashboard" to="" />
      </Route>
    </Route>
  </Route>,
);
