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
  query CheckDetailsContentQuery($ns: NamespaceInput!, $name: String!) {
    check(ns: $ns, name: $name) {
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
    const { match } = this.props;
    const ns = {
      organization: match.params.organization,
      environment: match.params.environment,
    };

    const { name } = match.params;

    return (
      <Query
        query={query}
        pollInterval={pollInterval}
        fetchPolicy="cache-and-network"
        variables={{ name, ns }}
      >
        {({
          aborted,
          client,
          data: { check } = {},
          loading,
          isPolling,
          refetch,
        }) => {
          if (!loading && !aborted && (!check || check.deleted)) {
            return <NotFound />;
          }

          return (
            <CheckDetailsContainer
              client={client}
              check={check}
              loading={(loading && !isPolling) || aborted}
              refetch={refetch}
            />
          );
        }}
      </Query>
    );
  }
}

export default CheckDetailsContent;
