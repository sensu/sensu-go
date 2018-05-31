import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";
import AppContent from "/components/AppContent";
import EventsContainer from "/components/EventsContainer";
import SearchBox from "/components/SearchBox";
import Content from "/components/Content";
import NotFoundView from "/components/views/NotFoundView";
import RefreshIcon from "@material-ui/icons/Refresh";
import ListToolbar from "/components/partials/ListToolbar";
import CollapsingMenu from "/components/CollapsingMenu";
import { withQueryParams } from "/components/QueryParams";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsContent extends React.Component {
  static propTypes = {
    match: PropTypes.object.isRequired,
    queryParams: PropTypes.shape({
      filter: PropTypes.string,
      order: PropTypes.string,
    }).isRequired,
    setQueryParams: PropTypes.func.isRequired,
  };

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String = "${defaultExpression}"
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
    return (
      <Query
        query={EventsContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...this.props.match.params, ...this.props.queryParams }}
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
                      placeholder="Filter events…"
                      initialValue={this.props.queryParams.filter}
                      onSearch={filter => this.props.setQueryParams({ filter })}
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
              <EventsContainer
                onQueryChange={this.props.setQueryParams}
                environment={environment}
                loading={loading || aborted}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withQueryParams(["filter", "order"])(EventsContent);
