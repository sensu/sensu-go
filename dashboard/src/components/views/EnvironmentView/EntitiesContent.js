import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Query from "/components/util/Query";
import Content from "/components/Content";
import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import EntitiesList from "/components/partials/EntitiesList";
import SearchBox from "/components/SearchBox";
import ListToolbar from "/components/partials/ListToolbar";
import LiveIcon from "/icons/Live";
import CollapsingMenu from "/components/partials/CollapsingMenu";
import { withQueryParams } from "/components/QueryParams";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

class EntitiesContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
    queryParams: PropTypes.shape({
      filter: PropTypes.string,
      order: PropTypes.string,
      offset: PropTypes.string,
      limit: PropTypes.string,
    }).isRequired,
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewEntitiesContentQuery(
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
      $order: EntityListOrder = ID
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EntitiesList_environment
      }
    }

    ${EntitiesList.fragments.environment}
  `;

  render() {
    const { queryParams, setQueryParams, match } = this.props;
    const { filter, order, limit = "50", offset = "0" } = queryParams;

    return (
      <Query
        query={EntitiesContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, filter, order, limit, offset }}
      >
        {({
          data: { environment } = {},
          loading,
          aborted,
          refetch,
          isPolling,
          startPolling,
          stopPolling,
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
                      placeholder="Filter entities…"
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
              <EntitiesList
                limit={limit}
                offset={offset}
                loading={(loading && !isPolling) || aborted}
                onChangeQuery={setQueryParams}
                environment={environment}
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
  EntitiesContent,
);
