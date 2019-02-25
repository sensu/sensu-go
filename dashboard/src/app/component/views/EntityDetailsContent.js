import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { FailedError } from "/lib/error/FetchError";

import EntityDetailsContainer from "/app/component/partial/EntityDetailsContainer";
import Loader from "/lib/component/base/Loader";
import NotFound from "/app/component/partial/NotFound";
import Query from "/lib/component/util/Query";

import { pollingDuration } from "/lib/constant/polling";

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
        onError={error => {
          if (error.networkError instanceof FailedError) {
            return;
          }

          throw error;
        }}
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
