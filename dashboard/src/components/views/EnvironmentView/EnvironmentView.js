import React from "react";
import PropTypes from "prop-types";
import { Switch, Route, Redirect } from "react-router-dom";

import AppWrapper from "/components/AppWrapper";
import LastEnvironmentUpdater from "/components/util/LastEnvironmentUpdater";
import NotFoundView from "/components/views/NotFoundView";

import ChecksContent from "./ChecksContent";
import EntitiesContent from "./EntitiesContent";
import EventsContent from "./EventsContent";
import CheckDetailsContent from "./CheckDetailsContent";
import EntityDetailsContent from "./EntityDetailsContent";
import SilencesContent from "./SilencesContent";
import EventDetailsContent from "./EventDetailsContent";

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
        <React.Fragment>
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
            <Route
              path={`${match.path}/checks/:name`}
              component={CheckDetailsContent}
            />
            <Redirect exact from={match.path} to={`${match.url}/events`} />
            <Route component={NotFoundView} />
          </Switch>
          <LastEnvironmentUpdater
            organization={match.params.organization}
            environment={match.params.environment}
          />
        </React.Fragment>
      </AppWrapper>
    );
  }
}

export default EnvironmentView;
