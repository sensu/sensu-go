import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import LiveIcon from "/icons/Live";

import { withQueryParams } from "/components/QueryParams";
import AppContent from "/components/AppContent";

import Query from "/components/util/Query";

import ChecksList from "/components/partials/ChecksList";
import ListToolbar from "/components/partials/ListToolbar";

import NotFoundView from "/components/views/NotFoundView";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import Content from "/components/Content";
import SearchBox from "/components/SearchBox";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

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
      $order: CheckListOrder
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...ChecksList_environment
      }
    }

    ${ChecksList.fragments.environment}
  `;

  render() {
    const { match, queryParams, setQueryParams } = this.props;

    const { limit = "50", offset = "0", order, filter } = queryParams;

    return (
      <Query
        query={ChecksContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, limit, offset, order, filter }}
      >
        {({
          aborted,
          data: { environment } = {},
          loading,
          isPolling,
          startPolling,
          stopPolling,
          refetch,
        }) => {
          if (!environment && !loading && !aborted) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Content gutters bottomMargin>
                <ListToolbar
                  renderSearch={
                    <SearchBox
                      placeholder="Filter checks…"
                      initialValue={filter}
                      onSearch={value => setQueryParams({ filter: value })}
                    />
                  }
                  renderMenuItems={
                    <CollapsingMenu.Button
                      title="LIVE"
                      icon={<LiveIcon active={isPolling} />}
                      onClick={() =>
                        isPolling ? stopPolling() : startPolling(pollInterval)
                      }
                    />
                  }
                />
              </Content>

              <ChecksList
                limit={limit}
                offset={offset}
                onChangeQuery={setQueryParams}
                environment={environment}
                loading={(loading && !isPolling) || aborted}
                refetch={refetch}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order", "offset", "limit"])(
  ChecksContent,
);
