import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { FailedError } from "/errors/FetchError";

import Query from "/components/util/Query";

import NotFound from "/components/partials/NotFound";
import Container from "/components/partials/EventDetailsContainer";

import { pollingDuration } from "/constants/polling";

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
        pollInterval={pollingDuration.short}
        variables={this.props.match.params}
        onError={error => {
          if (error.networkError instanceof FailedError) {
            return;
          }

          throw error;
        }}
      >
        {({ aborted, data: { event } = {}, networkStatus, refetch }) => {
          // see: https://github.com/apollographql/apollo-client/blob/master/packages/apollo-client/src/core/networkStatus.ts
          const loading = networkStatus < 6;

          if (!loading && !aborted && (!event || event.deleted)) {
            return <NotFound />;
          }

          return (
            <Container
              event={event}
              loading={loading || !!aborted}
              refetch={refetch}
            />
          );
        }}
      </Query>
    );
  }
}

export default EventDetailsContent;
