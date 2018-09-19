import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Query from "/components/util/Query";
import NotFound from "/components/partials/NotFound";
import EntitiesList from "/components/partials/EntitiesList";
import SearchBox from "/components/SearchBox";
import ListToolbar from "/components/partials/ListToolbar";
import LiveIcon from "/icons/Live";
import CollapsingMenu from "/components/partials/CollapsingMenu";
import { withQueryParams } from "/components/QueryParams";
import AppLayout from "/components/AppLayout";

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
    const { filter, order, limit = "25", offset = "0" } = queryParams;

    return (
      <Query
        query={EntitiesContent.query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
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
            return <NotFound />;
          }

          return (
            <div>
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
              <AppLayout.MobileFullWidthContent>
                <EntitiesList
                  limit={limit}
                  offset={offset}
                  loading={(loading && (!environment || !isPolling)) || aborted}
                  onChangeQuery={setQueryParams}
                  environment={environment}
                  refetch={refetch}
                />
              </AppLayout.MobileFullWidthContent>
            </div>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order", "offset", "limit"])(
  EntitiesContent,
);
