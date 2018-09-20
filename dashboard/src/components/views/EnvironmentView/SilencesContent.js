import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import withStateHandlers from "recompose/withStateHandlers";
import toRenderProps from "recompose/toRenderProps";

import AppLayout from "/components/AppLayout";
import Content from "/components/Content";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";
import SilencesList from "/components/partials/SilencesList";
import SilencesListToolbar from "/components/partials/SilencesList/SilencesListToolbar";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import { withQueryParams } from "/components/QueryParams";
import WithWidth from "/components/WithWidth";

const WithDialogState = toRenderProps(
  withStateHandlers(
    { isOpen: false },
    {
      open: () => () => ({ isOpen: true }),
      close: () => () => ({ isOpen: false }),
    },
  ),
);

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

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
      loading,
      aborted,
      refetch,
      poller,
    } = renderProps;

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
                loading={
                  (loading && (!namespace || !poller.isRunning())) || aborted
                }
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
    order: "ID",
  },
});
export default enhance(SilencesContent);
