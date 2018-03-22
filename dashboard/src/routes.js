import React from "react";
import { makeRouteConfig, Redirect, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import LoginPage from "./pages/Login";
import EventsPage from "./pages/EventsPage";
import ChecksPage from "./pages/ChecksPage";
import QueryPage from "./pages/GraphQLExplorerPage";
import RootRedirect from "./pages/RootRedirect";

import RestrictUnauthenticated from "./components/RestrictUnauthenticated";

export default makeRouteConfig(
  <Route>
    <Route path="/login" Component={LoginPage} />
    <Route path="query-explorer" Component={QueryPage} />
    <Route path="/" Component={RootRedirect} />

    <Route Component={RestrictUnauthenticated}>
      <Route
        path="/:organization/:environment"
        Component={AppWrapper}
        query={AppWrapper.query}
      >
        <Route
          path="events/"
          Component={EventsPage}
          query={EventsPage.query}
          prepareVariables={(params, route) => ({
            ...params,
            ...route.location.query,
          })}
        />
        <Redirect
          exact
          from="events"
          to="events?filter=HasCheck && HasIncident"
        />
        <Route path="checks" Component={ChecksPage} query={ChecksPage.query} />
        <Redirect from="dashboard" to="" />
      </Route>
    </Route>
  </Route>,
);
