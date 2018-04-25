import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import Paper from "material-ui/Paper";
import Button from "material-ui/Button";

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
      <Query query={ChecksContent.query} variables={match.params}>
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
