import React from "react";
import { makeRouteConfig, Redirect, Route } from "found";

import AppWrapper from "./components/AppWrapper";
import LoginPage from "./pages/Login";
import EventsPage from "./pages/EventsPage";
import ChecksPage from "./pages/ChecksPage";
import QueryPage from "./pages/GraphQLExplorerPage";

export default makeRouteConfig(
  <Route>
    <Route path="/login" Component={LoginPage} />
    <Redirect from="/" to="/default/default/" />

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
      <Route path="query-explorer" Component={QueryPage} />
      <Redirect from="dashboard" to="" />
    </Route>
  </Route>,
);
