import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";
import NotFound from "/components/partials/NotFound";
import CheckDetailsContainer from "/components/partials/CheckDetailsContainer";

// duration used when polling is enabled; set fairly high until we understand
// the impact.
const pollInterval = 1500; // 1.5s

const query = gql`
  query CheckDetailsContentQuery($namespace: String!, $name: String!) {
    check(namespace: $namespace, name: $name) {
      ...CheckDetailsContainer_check
    }
  }

  ${CheckDetailsContainer.fragments.check}
`;

class CheckDetailsContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    return (
      <Query
        query={query}
        pollInterval={pollInterval}
        fetchPolicy="cache-and-network"
        variables={this.props.match.params}
      >
        {({
          aborted,
          client,
          data: { check } = {},
          loading,
          poller,
          refetch,
        }) => {
          if (!loading && !aborted && (!check || check.deleted)) {
            return <NotFound />;
          }

          return (
            <CheckDetailsContainer
              client={client}
              check={check}
              loading={(loading && !poller.isRunning()) || aborted}
              refetch={refetch}
            />
          );
        }}
      </Query>
    );
  }
}

export default CheckDetailsContent;
