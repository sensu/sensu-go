import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { FailedError } from "/errors/FetchError";

import AppLayout from "/components/AppLayout";
import Content from "/components/Content";
import EntitiesList from "/components/partials/EntitiesList";
import ListToolbar from "/components/partials/EntitiesList/EntitiesListToolbar";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";
import { withQueryParams } from "/components/QueryParams";
import WithWidth from "/components/WithWidth";

import { pollingDuration } from "/constants/polling";

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
      $namespace: String!
      $limit: Int
      $offset: Int
      $order: EntityListOrder
      $filter: String
    ) {
      namespace(name: $namespace) {
        ...EntitiesList_namespace
      }
    }

    ${EntitiesList.fragments.namespace}
  `;

  renderContent = renderProps => {
    const { queryParams, setQueryParams } = this.props;
    const { filter, limit, offset, order } = queryParams;
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
            onClickReset={() => setQueryParams(q => q.reset())}
            query={filter}
          />
        </Content>

        <AppLayout.MobileFullWidthContent>
          <WithWidth>
            {({ width }) => (
              <EntitiesList
                editable={width !== "xs"}
                limit={limit}
                offset={offset}
                loading={(loading && !namespace) || aborted}
                onChangeQuery={setQueryParams}
                namespace={namespace}
                refetch={refetch}
                order={order}
              />
            )}
          </WithWidth>
        </AppLayout.MobileFullWidthContent>
      </div>
    );
  };

  render() {
    const { queryParams, match } = this.props;
    const variables = { ...match.params, ...queryParams };

    return (
      <Query
        query={EntitiesContent.query}
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
    order: "ID",
  },
});
export default enhance(EntitiesContent);
