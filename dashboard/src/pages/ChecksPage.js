import React from "react";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import { matchShape } from "found";

import Paper from "material-ui/Paper";
import AppContent from "../components/AppContent";
import CheckList from "../components/CheckList";

class CheckPage extends React.Component {
  static propTypes = {
    match: matchShape.isRequired,
  };

  static query = gql`
    query ChecksPageQuery($environment: String!, $organization: String!) {
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
        query={CheckPage.query}
        variables={match.params}
        // TODO: Replace polling with query subscription
        pollInterval={5000}
      >
        {({ data: { environment } = {}, loading, error }) => {
          // TODO: Connect this error handler to display a blocking error alert
          if (error) throw error;

          return (
            <AppContent>
              {!loading ? (
                <Paper>
                  <CheckList environment={environment} />
                </Paper>
              ) : (
                <div>Loading...</div>
              )}
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default CheckPage;
