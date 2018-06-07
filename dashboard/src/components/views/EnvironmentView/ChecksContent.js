import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Paper from "@material-ui/core/Paper";
import Button from "@material-ui/core/Button";

import { withQueryParams } from "/components/QueryParams";
import AppContent from "/components/AppContent";
import CheckList from "/components/CheckList";

import Query from "/components/util/Query";

import NotFoundView from "/components/views/NotFoundView";

class ChecksContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
    queryParams: PropTypes.shape({
      offset: PropTypes.string,
      limit: PropTypes.string,
    }).isRequired,
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewChecksContentQuery(
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
    ) {
      environment(organization: $organization, environment: $environment) {
        ...CheckList_environment
      }
    }

    ${CheckList.fragments.environment}
  `;

  render() {
    const { match, queryParams, setQueryParams } = this.props;

    const { limit = "50", offset = "0" } = queryParams;

    return (
      <Query
        query={ChecksContent.query}
        variables={{ ...match.params, limit, offset }}
      >
        {({ data: { environment } = {}, loading, aborted, refetch }) => {
          if (!environment && !loading && !aborted) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Button onClick={() => refetch()}>reload</Button>
              <Paper>
                <CheckList
                  limit={limit}
                  offset={offset}
                  onChangeParams={setQueryParams}
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

export default withQueryParams(["offset", "limit"])(ChecksContent);
