import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Paper from "@material-ui/core/Paper";
import Button from "@material-ui/core/Button";

import Query from "/components/util/Query";

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
        {({ data: { environment } = {}, loading, aborted, refetch }) => {
          if (!environment && !loading && !aborted) return <NotFoundView />;

          return (
            <AppContent>
              <Button onClick={() => refetch()}>reload</Button>
              <Paper>
                <CheckList
                  environment={environment}
                  loading={loading || aborted}
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
