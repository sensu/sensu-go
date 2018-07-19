import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import AppContent from "/components/AppContent";
import EntityDetailsContainer from "/components/partials/EntityDetailsContainer";
import Loader from "/components/util/Loader";
import NotFoundView from "/components/views/NotFoundView";
import Query from "/components/util/Query";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 1500; // 1.5s

const query = gql`
  query EntityDetailsContentQuery($ns: NamespaceInput!, $name: String!) {
    entity(ns: $ns, name: $name) {
      deleted @client
      ...EntityDetailsContainer_entity
    }
  }

  ${EntityDetailsContainer.fragments.entity}
`;

class EntityDetailsContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    const { match } = this.props;
    const { organization, environment, ...params } = match.params;
    const ns = { organization, environment };

    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={{ ...params, ns }}
      >
        {({
          data: { entity } = {},
          loading,
          aborted,
          isPolling,
          startPolling,
          stopPolling,
        }) => {
          if (!loading && !aborted && (!entity || entity.deleted)) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Loader loading={(loading && !isPolling) || aborted} passthrough>
                {entity && (
                  <EntityDetailsContainer
                    entity={entity}
                    poller={{
                      running: isPolling,
                      start: startPolling,
                      stop: stopPolling,
                    }}
                  />
                )}
              </Loader>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EntityDetailsContent;
