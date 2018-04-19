import React from "react";
import PropTypes from "prop-types";
import { Switch, Route, Redirect } from "react-router-dom";

import AppWrapper from "../../AppWrapper";

import NotFoundView from "../../views/NotFoundView";

import DashboardContent from "./DashboardContent";
import ChecksContent from "./ChecksContent";
import EventsContent from "./EventsContent";

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
          <Route exact path={match.path} component={DashboardContent} />
          <Route
            exact
            path={`${match.path}/checks`}
            component={ChecksContent}
          />
          <Route
            exact
            path={`${match.path}/events`}
            component={EventsContent}
          />
          <Redirect exact from={`${match.path}/dashboard`} to={match.path} />
          <Route component={NotFoundView} />
        </Switch>
      </AppWrapper>
    );
  }
}

export default EnvironmentView;
