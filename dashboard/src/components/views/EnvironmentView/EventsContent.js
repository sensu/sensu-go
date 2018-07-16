import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import RefreshIcon from "@material-ui/icons/Refresh";

import Query from "/components/util/Query";

import EventsList from "/components/partials/EventsList";
import ListToolbar from "/components/partials/ListToolbar";

import NotFoundView from "/components/views/NotFoundView";

import SearchBox from "/components/SearchBox";
import { withQueryParams } from "/components/QueryParams";
import AppContent from "/components/AppContent";
import CollapsingMenu from "/components/CollapsingMenu";
import Content from "/components/Content";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

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
      $order: EventsListOrder = SEVERITY
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

    const { filter, order, limit = "50", offset = "0" } = queryParams;

    return (
      <Query
        query={EventsContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, filter, order, limit, offset }}
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
                      placeholder="Filter events…"
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
              <EventsList
                limit={limit}
                offset={offset}
                onChangeQuery={setQueryParams}
                environment={environment}
                loading={loading || aborted}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order", "offset", "limit"])(
  EventsContent,
);
