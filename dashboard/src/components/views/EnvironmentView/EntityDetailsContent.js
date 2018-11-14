import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import EntityDetailsContainer from "/components/partials/EntityDetailsContainer";
import Loader from "/components/util/Loader";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 1500; // 1.5s

const query = gql`
  query EntityDetailsContentQuery($namespace: String!, $name: String!) {
    entity(namespace: $namespace, name: $name) {
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
    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        pollInterval={pollInterval}
        variables={this.props.match.params}
      >
        {({ data: { entity } = {}, loading, aborted, poller, refetch }) => {
          if (!loading && !aborted && (!entity || entity.deleted)) {
            return <NotFound />;
          }

          return (
            <Loader
              loading={(loading && !poller.isRunning()) || aborted}
              passthrough
            >
              {entity && (
                <EntityDetailsContainer entity={entity} refetch={refetch} />
              )}
            </Loader>
          );
        }}
      </Query>
    );
  }
}

export default EntityDetailsContent;
