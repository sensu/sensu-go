import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";

import NotFound from "/components/partials/NotFound";
import Container from "/components/partials/EventDetailsContainer";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 1500; // 1.5s

const query = gql`
  query EventDetailsContentQuery(
    $namespace: String!
    $check: String!
    $entity: String!
  ) {
    event(namespace: $namespace, entity: $entity, check: $check) {
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
    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={this.props.match.params}
      >
        {({ data: { event } = {}, networkStatus, aborted }) => {
          // see: https://github.com/apollographql/apollo-client/blob/master/packages/apollo-client/src/core/networkStatus.ts
          const loading = networkStatus < 6;

          if (!loading && !aborted && (!event || event.deleted)) {
            return <NotFound />;
          }

          return <Container event={event} loading={loading || !!aborted} />;
        }}
      </Query>
    );
  }
}

export default EventDetailsContent;
