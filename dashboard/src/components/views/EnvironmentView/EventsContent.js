import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import AppContent from "/components/AppContent";
import EventsContainer from "/components/EventsContainer";
import SearchBox from "/components/SearchBox";
import Content from "/components/Content";
import NotFoundView from "/components/views/NotFoundView";
import RefreshIcon from "@material-ui/icons/Refresh";
import WithQueryParams from "/components/WithQueryParams";
import ListToolbar from "/components/partials/ListToolbar";
import { CollapsingMenuItem } from "/components/CollapsingMenu";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String!
      $order: EventsListOrder = SEVERITY
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsContainer_environment
      }
    }

    ${EventsContainer.fragments.environment}
  `;

  render() {
    const { classes, match } = this.props;

    return (
      <WithQueryParams>
        {(query, setQuery) => (
          <Query
            query={EventsContent.query}
            fetchPolicy="cache-and-network"
            variables={{
              ...match.params,
              filter: query.get("filter") || defaultExpression,
              order: query.get("order"),
            }}
          >
            {({ data: { environment } = {}, loading, error, refetch }) => {
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
                          placeholder="Filter events…"
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
                  <EventsContainer
                    onQueryChange={setQuery}
                    environment={environment}
                    loading={loading}
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

export default EventsContent;
