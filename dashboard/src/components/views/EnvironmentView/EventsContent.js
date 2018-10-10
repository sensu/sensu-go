import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import AppLayout from "/components/AppLayout";
import Content from "/components/Content";
import EventsList from "/components/partials/EventsList";
import ListToolbar from "/components/partials/EventsList/EventsListToolbar";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";
import { withQueryParams } from "/components/QueryParams";

// If none given default expression is used.
const defaultExpression = "HasCheck";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

class EventsContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,

    // from withQueryParams HOC
    queryParams: PropTypes.shape({
      filter: PropTypes.string,
      order: PropTypes.string,
      offset: PropTypes.string,
      limit: PropTypes.string,
    }).isRequired,

    // from withQueryParams HOC
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String = "${defaultExpression}"
      $order: EventsListOrder
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

  renderContent = renderProps => {
    const { queryParams, setQueryParams } = this.props;
    const { filter, limit, offset } = queryParams;
    const {
      data: { environment } = {},
      loading,
      aborted,
      isPolling,
      refetch,
    } = renderProps;

    if (!environment && !loading && !aborted) {
      return <NotFound />;
    }

    return (
      <div>
        <Content marginBottom>
          <ListToolbar
            onChangeQuery={value => setQueryParams({ filter: value })}
            onClickReset={() =>
              setQueryParams(q => q.reset(["filter", "order"]))
            }
            query={filter}
          />
        </Content>

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
  };

  render() {
    const { queryParams, match } = this.props;
    const variables = { ...match.params, ...queryParams };

    return (
      <Query
        query={EventsContent.query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={variables}
      >
        {this.renderContent}
      </Query>
    );
  }
}

const enhance = withQueryParams({
  keys: ["filter", "order", "offset", "limit"],
  defaults: {
    limit: "25",
    offset: "0",
    order: "LASTOK",
  },
});
export default enhance(EventsContent);
