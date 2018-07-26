import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";

import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import Container from "/components/partials/EventDetailsContainer";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 1500; // 1.5s

const query = gql`
  query EventDetailsContentQuery(
    $ns: NamespaceInput!
    $check: String!
    $entity: String!
  ) {
    event(ns: $ns, entity: $entity, check: $check) {
      deleted @client
      ...EventDetailsContainer_event
    }
  }

  ${Container.fragments.event}
`;

class EventDetailsContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    const { match } = this.props;
    const ns = {
      organization: match.params.organization,
      environment: match.params.environment,
    };

    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={{ ...match.params, ns }}
      >
        {({
          client,
          data: { event } = {},
          loading,
          aborted,
          isPolling,
          startPolling,
          stopPolling,
        }) => {
          if (!loading && !aborted && (!event || event.deleted)) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Container
                client={client}
                event={event}
                loading={(loading && !isPolling) || aborted}
                poller={{
                  running: isPolling,
                  start: startPolling,
                  stop: stopPolling,
                }}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EventDetailsContent;
