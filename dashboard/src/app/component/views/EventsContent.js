import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { FailedError } from "/lib/error/FetchError";

import AppLayout from "/lib/component/base/AppLayout";
import Content from "/lib/component/base/Content";
import Query from "/lib/component/util/Query";
import { withQueryParams } from "/lib/component/util/QueryParams";
import ToastConnector from "/lib/component/relocation/ToastConnector";
import WithWidth from "/lib/component/util/WithWidth";

import { pollingDuration } from "/lib/constant/polling";

import EventsList from "/app/component/partial/EventsList";
import ListToolbar from "/app/component/partial/EventsList/EventsListToolbar";
import NotFound from "/app/component/partial/NotFound";

// If none given default expression is used.
const defaultExpression = "has_check";

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
      $namespace: String!
      $filter: String = "${defaultExpression}"
      $order: EventsListOrder
      $limit: Int,
      $offset: Int,
    ) {
      namespace(name: $namespace) {
        ...EventsList_namespace
      }
    }

    ${EventsList.fragments.namespace}
  `;

  renderContent = renderProps => {
    const { queryParams, setQueryParams } = this.props;
    const { filter, limit, offset } = queryParams;
    const {
      data: { namespace } = {},
      networkStatus,
      aborted,
      refetch,
    } = renderProps;

    // see: https://github.com/apollographql/apollo-client/blob/master/packages/apollo-client/src/core/networkStatus.ts
    const loading = networkStatus < 6;

    if (!namespace && !loading && !aborted) {
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
          <ToastConnector>
            {({ addToast }) => (
              <WithWidth>
                {({ width }) => (
                  <EventsList
                    addToast={addToast}
                    editable={width !== "xs"}
                    limit={limit}
                    offset={offset}
                    onChangeQuery={setQueryParams}
                    namespace={namespace}
                    loading={(loading && !namespace) || aborted}
                    refetch={refetch}
                  />
                )}
              </WithWidth>
            )}
          </ToastConnector>
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
        pollInterval={pollingDuration.short}
        variables={variables}
        onError={error => {
          if (error.networkError instanceof FailedError) {
            return;
          }

          throw error;
        }}
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
