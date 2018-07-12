import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import RefreshIcon from "@material-ui/icons/Refresh";

import { withQueryParams } from "/components/QueryParams";
import AppContent from "/components/AppContent";

import Query from "/components/util/Query";

import ChecksList from "/components/partials/ChecksList";
import ListToolbar from "/components/partials/ListToolbar";

import NotFoundView from "/components/views/NotFoundView";

import CollapsingMenu from "/components/CollapsingMenu";
import Content from "/components/Content";
import SearchBox from "/components/SearchBox";

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

  render() {
    const { match, queryParams, setQueryParams } = this.props;

    const { limit = "50", offset = "0", order, filter } = queryParams;

    return (
      <Query
        query={ChecksContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, limit, offset, order, filter }}
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
                      placeholder="Filter checks…"
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

              <ChecksList
                limit={limit}
                offset={offset}
                onChangeQuery={setQueryParams}
                environment={environment}
                loading={loading || aborted}
                refetch={refetch}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order", "offset", "limit"])(
  ChecksContent,
);
