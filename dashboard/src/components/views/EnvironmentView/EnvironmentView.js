import React from "react";
import PropTypes from "prop-types";
import { Switch, Route, Redirect } from "react-router-dom";

import AppWrapper from "/components/AppWrapper";
import NotFoundView from "/components/views/NotFoundView";

import ChecksContent from "./ChecksContent";
import EntitiesContent from "./EntitiesContent";
import EventsContent from "./EventsContent";
import EventDetailsContent from "./EventDetailsContent";
import EntityDetailsContent from "./EntityDetailsContent";
import SilencesContent from "./SilencesContent";

class EnvironmentView extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    const { match } = this.props;

    return (
      <AppWrapper
        organization={match.params.organization}
        environment={match.params.environment}
      >
        <Switch>
          <Route
            exact
            path={`${match.path}/checks`}
            component={ChecksContent}
          />
          <Route
            exact
            path={`${match.path}/entities`}
            component={EntitiesContent}
          />
          <Route
            exact
            path={`${match.path}/events`}
            component={EventsContent}
          />
          <Route
            path={`${match.path}/events/:entity/:check`}
            component={EventDetailsContent}
          />
          <Route
            path={`${match.path}/entities/:name`}
            component={EntityDetailsContent}
          />
          <Route
            exact
            path={`${match.path}/silences`}
            component={SilencesContent}
          />
          <Redirect exact from={match.path} to={`${match.url}/events`} />
          <Route component={NotFoundView} />
        </Switch>
      </AppWrapper>
    );
  }
}

export default EnvironmentView;
