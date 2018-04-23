import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import Paper from "material-ui/Paper";

import AppContent from "/components/AppContent";
import CheckList from "/components/CheckList";

import NotFoundView from "/components/views/NotFoundView";

class ChecksContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewChecksContentQuery(
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...CheckList_environment
      }
    }

    ${CheckList.fragments.environment}
  `;

  render() {
    const { match } = this.props;

    return (
      <Query
        query={ChecksContent.query}
        variables={match.params}
        // TODO: Replace polling with query subscription
        pollInterval={5000}
      >
        {({ data: { environment } = {}, loading, error }) => {
          // TODO: Connect this error handler to display a blocking error alert
          if (error) throw error;

          if (!environment && !loading) return <NotFoundView />;

          return (
            <AppContent>
              <Paper>
                <CheckList environment={environment} loading={loading} />
              </Paper>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default ChecksContent;
