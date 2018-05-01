import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { Query } from "react-apollo";
import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import Container from "/components/partials/EventDetailsContainer";

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
        variables={{ ...match.params, ns }}
      >
        {({ client, data: { event } = {}, loading }) => {
          if (!loading && (!event || event.deleted)) return <NotFoundView />;
          return (
            <AppContent>
              <Container client={client} event={event} loading={loading} />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EventDetailsContent;
