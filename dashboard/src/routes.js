import React from "react";
import { makeRouteConfig, Redirect, Route } from "found";

import AppRoot from "./components/AppRoot";
import AppWrapper from "./components/AppWrapper";
import RestrictUnauthenticated from "./components/RestrictUnauthenticated";

import LoginPage from "./pages/Login";
import EventsPage from "./pages/EventsPage";
import ChecksPage from "./pages/ChecksPage";
import QueryPage from "./pages/GraphQLExplorerPage";
import RootRedirect from "./pages/RootRedirect";

export default makeRouteConfig(
  <Route Component={AppRoot}>
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
          path="events"
          Component={EventsPage}
          query={EventsPage.query}
          prepareVariables={(params, route) => ({
            ...params,
            ...route.location.query,
          })}
        />
        <Route path="checks" Component={ChecksPage} query={ChecksPage.query} />
        <Redirect from="dashboard" to="" />
      </Route>
    </Route>
  </Route>,
);
