import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import LiveIcon from "/icons/Live";
import AppLayout from "/components/AppLayout";

import Query from "/components/util/Query";

import EventsList from "/components/partials/EventsList";
import ListToolbar from "/components/partials/ListToolbar";

import NotFoundView from "/components/views/NotFoundView";

import SearchBox from "/components/SearchBox";
import { withQueryParams } from "/components/QueryParams";
import CollapsingMenu from "/components/partials/CollapsingMenu";

// If none given default expression is used.
const defaultExpression = "HasCheck";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 5000; // 5s

class EventsContent extends React.Component {
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
    query EnvironmentViewEventsContentQuery(
      $filter: String = "${defaultExpression}"
      $order: EventsListOrder = LASTOK
      $limit: Int,
      $offset: Int,
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsList_environment
      }
    }

    ${EventsList.fragments.environment}
  `;

  render() {
    const { queryParams, setQueryParams, match } = this.props;
    const { filter, order, limit = "25", offset = "0" } = queryParams;

    return (
      <Query
        query={EventsContent.query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={{ ...match.params, filter, order, limit, offset }}
      >
        {({
          data: { environment } = {},
          loading,
          aborted,
          isPolling,
          startPolling,
          stopPolling,
          refetch,
        }) => {
          if (!environment && !loading && !aborted) {
            return <NotFoundView />;
          }

          return (
            <div>
              <ListToolbar
                renderSearch={
                  <SearchBox
                    placeholder="Filter events…"
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
                <EventsList
                  limit={limit}
                  offset={offset}
                  onChangeQuery={setQueryParams}
                  environment={environment}
                  loading={(loading && (!environment || !isPolling)) || aborted}
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
  EventsContent,
);
