import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import EntityDetailsContainer from "/components/partials/EntityDetailsContainer";
import Loader from "/components/util/Loader";
import NotFound from "/components/partials/NotFound";
import Query from "/components/util/Query";

import { pollingDuration } from "/constants/polling";

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
        pollInterval={pollingDuration.short}
        variables={this.props.match.params}
      >
        {({ data: { entity } = {}, networkStatus, aborted, refetch }) => {
          // see: https://github.com/apollographql/apollo-client/blob/master/packages/apollo-client/src/core/networkStatus.ts
          const loading = networkStatus < 6;

          if (!loading && !aborted && (!entity || entity.deleted)) {
            return <NotFound />;
          }

          return (
            <Loader loading={loading || aborted} passthrough>
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
