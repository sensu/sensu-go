import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import AppLayout from "/components/AppLayout";
import ChecksList from "/components/partials/ChecksList";
import Content from "/components/Content";
import ListToolbar from "/components/partials/ChecksList/ChecksListToolbar";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";
import ToastConnector from "/components/relocation/ToastConnector";
import { withQueryParams } from "/components/QueryParams";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 2500; // 2.5s

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
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
      $order: CheckListOrder
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...ChecksList_environment
      }
    }

    ${ChecksList.fragments.environment}
  `;

  renderContent = renderProps => {
    const { queryParams, setQueryParams } = this.props;
    const { limit, offset, filter } = queryParams;
    const {
      aborted,
      data: { environment } = {},
      loading,
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
              <ChecksList
                limit={limit}
                offset={offset}
                onChangeQuery={setQueryParams}
                environment={environment}
                loading={(loading && (!environment || !isPolling)) || aborted}
                refetch={refetch}
                order={queryParams.order}
                addToast={addToast}
              />
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
    order: "NAME",
  },
});
export default enhance(ChecksContent);
