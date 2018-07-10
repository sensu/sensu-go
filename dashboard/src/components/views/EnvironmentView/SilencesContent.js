import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import AppContent from "/components/AppContent";
import CollapsingMenu from "/components/CollapsingMenu";
import Content from "/components/Content";
import ListToolbar from "/components/partials/ListToolbar";
import NotFoundView from "/components/views/NotFoundView";
import Paper from "@material-ui/core/Paper";
import Query from "/components/util/Query";
import RefreshIcon from "@material-ui/icons/Refresh";
import SearchBox from "/components/SearchBox";
import SilencesList from "/components/partials/SilencesList";
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
              <Content gutters bottomMargin>
                <ListToolbar
                  renderSearch={
                    <SearchBox
                      placeholder="Filter entities…"
                      initialValue={filter}
                      onSearch={value => setQueryParams({ filter: value })}
                    />
                  }
                  renderMenuItems={
                    <CollapsingMenu.Button
                      title="Reload"
                      icon={<RefreshIcon />}
                      onClick={() => refetch()}
                    />
                  }
                />
              </Content>
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
