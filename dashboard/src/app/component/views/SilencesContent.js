import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import withStateHandlers from "recompose/withStateHandlers";
import toRenderProps from "recompose/toRenderProps";

import { FailedError } from "/lib/error/FetchError";

import AppLayout from "/lib/component/base/AppLayout";
import Content from "/lib/component/base/Content";
import Query from "/lib/component/util/Query";
import { withQueryParams } from "/lib/component/util/QueryParams";
import WithWidth from "/lib/component/util/WithWidth";

import { pollingDuration } from "/lib/constant/polling";

import NotFound from "/app/component/partial/NotFound";
import SilencesList from "/app/component/partial/SilencesList";
import SilencesListToolbar from "/app/component/partial/SilencesList/SilencesListToolbar";
import SilenceEntryDialog from "/app/component/partial/SilenceEntryDialog";

const WithDialogState = toRenderProps(
  withStateHandlers(
    { isOpen: false },
    {
      open: () => () => ({ isOpen: true }),
      close: () => () => ({ isOpen: false }),
    },
  ),
);

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
      $namespace: String!
      $limit: Int
      $offset: Int
      $order: SilencesListOrder
      $filter: String
    ) {
      namespace(name: $namespace) {
        ...SilencesList_namespace
      }
    }

    ${SilencesList.fragments.namespace}
  `;

  renderContent = renderProps => {
    const { match, queryParams, setQueryParams } = this.props;
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
        <WithDialogState>
          {newDialog => (
            <React.Fragment>
              <Content marginBottom>
                <SilencesListToolbar
                  filter={filter}
                  onChangeQuery={val => setQueryParams({ filter: val })}
                  onClickCreate={newDialog.open}
                  onClickReset={() =>
                    setQueryParams(q => q.reset(["filter", "offset"]))
                  }
                />
              </Content>

              {newDialog.isOpen && (
                <SilenceEntryDialog
                  values={{
                    namespace: match.params.namespace,
                    props: {},
                  }}
                  onClose={() => {
                    // TODO: Only refetch / poison list on success
                    refetch();
                    newDialog.close();
                  }}
                />
              )}
            </React.Fragment>
          )}
        </WithDialogState>

        <AppLayout.MobileFullWidthContent>
          <WithWidth>
            {({ width }) => (
              <SilencesList
                editable={width !== "xs"}
                limit={limit}
                offset={offset}
                order={order}
                onChangeQuery={setQueryParams}
                namespace={namespace}
                loading={(loading && !namespace) || aborted}
                refetch={refetch}
              />
            )}
          </WithWidth>
        </AppLayout.MobileFullWidthContent>
      </div>
    );
  };

  render() {
    const { match, queryParams } = this.props;
    const variables = { ...match.params, ...queryParams };

    return (
      <Query
        query={SilencesContent.query}
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
export default enhance(SilencesContent);
