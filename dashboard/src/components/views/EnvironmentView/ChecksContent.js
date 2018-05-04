import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import Paper from "material-ui/Paper";
import Button from "material-ui/Button";

import AppContent from "/components/AppContent";
import CheckList from "/components/CheckList";

import NotFoundView from "/components/views/NotFoundView";

// Hardcoded page size
const fetchLimit = 100;

class ChecksContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewChecksContentQuery(
      $environment: String!
      $organization: String!
      $limit: Int!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...CheckList_environment
      }
    }

    ${CheckList.fragments.environment}
  `;

  render() {
    const { match } = this.props;
    const variables = { limit: fetchLimit, ...match.params };

    return (
      <Query query={ChecksContent.query} variables={variables}>
        {({ data: { environment } = {}, loading, error, refetch }) => {
          // TODO: Connect this error handler to display a blocking error alert
          if (error) throw error;

          if (!environment && !loading) return <NotFoundView />;

          return (
            <AppContent>
              <Button onClick={() => refetch()}>reload</Button>
              <Paper>
                <CheckList
                  environment={environment}
                  loading={loading}
                  refetch={refetch}
                />
              </Paper>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default ChecksContent;
