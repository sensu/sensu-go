import React from "react";
import PropTypes from "prop-types";
import { Switch, Route, Redirect } from "react-router-dom";
import gql from "graphql-tag";

import { FailedError } from "/errors/FetchError";

import AppBar from "/components/AppBar";
import AppLayout from "/components/AppLayout";
import QuickNav from "/components/QuickNav";
import Loader from "/components/util/Loader";
import LastNamespaceUpdater from "/components/util/LastNamespaceUpdater";
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
    query EnvironmentViewQuery($namespace: String!) {
      viewer {
        ...AppBar_viewer
      }

      namespace(name: $namespace) {
        ...AppBar_namespace
      }
    }

    ${AppBar.fragments.viewer}
    ${AppBar.fragments.namespace}
  `;

  render() {
    const { match } = this.props;
    const namespaceParam = match.params.namespace;

    return (
      <React.Fragment>
        <Query
          query={EnvironmentView.query}
          variables={{ namespace: namespaceParam }}
          onError={error => {
            if (error.networkError instanceof FailedError) {
              return;
            }

            throw error;
          }}
        >
          {({ data: { viewer, namespace } = {}, loading, aborted }) => (
            <Loader loading={loading}>
              <AppLayout
                topBar={
                  <AppBar
                    loading={loading || aborted}
                    namespace={namespace}
                    viewer={viewer}
                  />
                }
                quickNav={<QuickNav namespace={namespaceParam} />}
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

        <LastNamespaceUpdater namespace={namespaceParam} />
      </React.Fragment>
    );
  }
}

export default EnvironmentView;
