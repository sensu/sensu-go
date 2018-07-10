import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Paper from "@material-ui/core/Paper";
import Button from "@material-ui/core/Button";
import AppContent from "/components/AppContent";
import SilencesList from "/components/partials/SilencesList";
import Query from "/components/util/Query";
import NotFoundView from "/components/views/NotFoundView";
import { withQueryParams } from "/components/QueryParams";

class SilencesContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
    queryParams: PropTypes.shape({
      offset: PropTypes.string,
      limit: PropTypes.string,
    }).isRequired,
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewSilencesContentQuery(
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
      $order: SilencesListOrder
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...SilencesList_environment
      }
    }

    ${SilencesList.fragments.environment}
  `;

  render() {
    const { match, queryParams, setQueryParams } = this.props;
    const { limit = "50", offset = "0", order, filter } = queryParams;

    return (
      <Query
        query={SilencesContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, limit, offset, order, filter }}
      >
        {({ data: { environment } = {}, loading, aborted, refetch }) => {
          if (!environment && !loading && !aborted) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Button onClick={() => refetch()}>reload</Button>
              <Paper>
                <SilencesList
                  limit={limit}
                  offset={offset}
                  onChangeQuery={setQueryParams}
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

export default withQueryParams(["filter", "order", "offset", "limit"])(
  SilencesContent,
);
