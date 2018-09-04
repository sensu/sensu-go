import React from "react";
import PropTypes from "prop-types";
import { Switch, Route, Redirect } from "react-router-dom";
import gql from "graphql-tag";

import AppBar from "/components/AppBar";
import AppLayout from "/components/AppLayout";
import QuickNav from "/components/QuickNav";
import Loader from "/components/util/Loader";
import LastEnvironmentUpdater from "/components/util/LastEnvironmentUpdater";
import NotFoundView from "/components/views/NotFoundView";
import Query from "/components/util/Query";

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

  static query = gql`
    query EnvironmentViewQuery($environment: String!, $organization: String!) {
      viewer {
        ...AppBar_viewer
      }

      environment(organization: $organization, environment: $environment) {
        ...AppBar_environment
      }
    }

    ${AppBar.fragments.viewer}
    ${AppBar.fragments.environment}
  `;

  render() {
    const { match } = this.props;

    return (
      <React.Fragment>
        <Query
          query={EnvironmentView.query}
          variables={{
            organization: match.params.organization,
            environment: match.params.environment,
          }}
        >
          {({ data: { viewer, environment } = {}, loading, aborted }) => (
            <Loader loading={loading}>
              <AppLayout
                topBar={
                  <AppBar
                    loading={loading || aborted}
                    environment={environment}
                    viewer={viewer}
                  />
                }
                quickNav={
                  <QuickNav organization={"default"} environment={"default"} />
                }
                content={
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
                    <Redirect
                      exact
                      from={match.path}
                      to={`${match.url}/events`}
                    />
                    <Route component={NotFoundView} />
                  </Switch>
                }
              />
            </Loader>
          )}
        </Query>
        <LastEnvironmentUpdater
          organization={match.params.organization}
          environment={match.params.environment}
        />
      </React.Fragment>
    );
  }
}

export default EnvironmentView;
