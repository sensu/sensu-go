import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Query from "/components/util/Query";
import Content from "/components/Content";
import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import EntitiesList from "/components/partials/EntitiesList";
import SearchBox from "/components/SearchBox";
import ListToolbar from "/components/partials/ListToolbar";
import RefreshIcon from "@material-ui/icons/Refresh";
import CollapsingMenu from "/components/CollapsingMenu";
import { withQueryParams } from "/components/QueryParams";

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
      $environment: String!
      $organization: String!
      $limit: Int
      $offset: Int
      $order: EntityListOrder = ID
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EntitiesList_environment
      }
    }

    ${EntitiesList.fragments.environment}
  `;

  render() {
    const { queryParams, setQueryParams, match } = this.props;

    return (
      <Query
        query={EntitiesContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, ...queryParams }}
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
                      placeholder="Filter entities…"
                      initialValue={queryParams.filter}
                      onSearch={filter => setQueryParams({ filter })}
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
              <EntitiesList
                limit={queryParams.limit}
                offset={queryParams.offset}
                onChangeParams={setQueryParams}
                loading={loading || aborted}
                environment={environment}
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
  EntitiesContent,
);
