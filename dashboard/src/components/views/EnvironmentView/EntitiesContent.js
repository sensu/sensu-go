import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import Content from "/components/Content";
import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import EntitiesList from "/components/partials/EntitiesList";
import WithQueryParams from "/components/WithQueryParams";
import SearchBox from "/components/SearchBox";
import ListToolbar from "/components/partials/ListToolbar";
import RefreshIcon from "@material-ui/icons/Refresh";
import { CollapsingMenuItem } from "/components/CollapsingMenu";

class EntitiesContent extends React.PureComponent {
  static propTypes = {
    // eslint-disable-next-line react/no-unused-prop-types
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewEntitiesContentQuery(
      $environment: String!
      $organization: String!
      $sort: EntityListOrder = ID
      $filter: String
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EntitiesList_environment
      }
    }

    ${EntitiesList.fragments.environment}
  `;

  render() {
    return (
      <WithQueryParams>
        {(query, setQuery) => (
          <Query
            query={EntitiesContent.query}
            fetchPolicy="cache-and-network"
            variables={{
              ...this.props.match.params,
              filter: query.get("filter"),
              order: query.get("order"),
            }}
          >
            {({ data: { environment } = {}, loading, error, refetch }) => {
              // TODO: Connect this error handler to display a blocking error alert
              if (error) throw error;

              if (!environment && !loading) return <NotFoundView />;

              return (
                <AppContent>
                  <Content gutters bottomMargin>
                    <ListToolbar
                      renderSearch={
                        <SearchBox
                          onSearch={val => setQuery("filter", val)}
                          initialValue={query.get("filter")}
                          placeholder="Filter entities…"
                        />
                      }
                      renderMenuItems={
                        <CollapsingMenuItem
                          title="Reload"
                          icon={<RefreshIcon />}
                          onClick={() => refetch()}
                        />
                      }
                    />
                  </Content>
                  <EntitiesList
                    loading={loading}
                    environment={environment}
                    refetch={refetch}
                  />
                </AppContent>
              );
            }}
          </Query>
        )}
      </WithQueryParams>
    );
  }
}

export default EntitiesContent;
