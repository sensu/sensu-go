import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { FailedError } from "/lib/error/FetchError";
import AppLayout from "/lib/component/base/AppLayout";
import Content from "/lib/component/base/Content";
import Query from "/lib/component/util/Query";
import ToastConnector from "/lib/component/relocation/ToastConnector";
import { withQueryParams } from "/lib/component/util/QueryParams";
import WithWidth from "/lib/component/util/WithWidth";
import { pollingDuration } from "/lib/constant/polling";

import ChecksList from "/app/component/partial/ChecksList";
import ListToolbar from "/app/component/partial/ChecksList/ChecksListToolbar";
import NotFound from "/app/component/partial/NotFound";

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
      $namespace: String!
      $limit: Int
      $offset: Int
      $order: CheckListOrder
      $filter: String
    ) {
      namespace(name: $namespace) {
        ...ChecksList_namespace
      }
    }

    ${ChecksList.fragments.namespace}
  `;

  renderContent = renderProps => {
    const { queryParams, setQueryParams } = this.props;
    const { limit, offset, filter } = queryParams;
    const {
      aborted,
      data: { namespace } = {},
      networkStatus,
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
            query={filter}
            onChangeQuery={value => setQueryParams({ filter: value })}
            onClickReset={() =>
              setQueryParams(q => q.reset(["filter", "order"]))
            }
          />
        </Content>

        <AppLayout.MobileFullWidthContent>
          <ToastConnector>
            {({ addToast }) => (
              <WithWidth>
                {({ width }) => (
                  <ChecksList
                    editable={width !== "xs"}
                    limit={limit}
                    offset={offset}
                    onChangeQuery={setQueryParams}
                    namespace={namespace}
                    loading={(loading && !namespace) || aborted}
                    refetch={refetch}
                    order={queryParams.order}
                    addToast={addToast}
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
    const { match, queryParams } = this.props;
    const variables = { ...match.params, ...queryParams };

    return (
      <Query
        query={ChecksContent.query}
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
    order: "NAME",
  },
});
export default enhance(ChecksContent);
