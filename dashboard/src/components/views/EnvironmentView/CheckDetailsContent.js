import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";

import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import CheckDetailsContainer from "/components/partials/CheckDetailsContainer";

const query = gql`
  query CheckDetailsContentQuery($ns: NamespaceInput!, $name: String!) {
    check(ns: $ns, name: $name) {
      ...CheckDetailsContainer_checkConfig
    }
  }

  ${CheckDetailsContainer.fragments.checkConfig}
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
        fetchPolicy="cache-and-network"
        variables={{ name, ns }}
      >
        {({ client, data: { check } = {}, loading, aborted }) => {
          if (!loading && !aborted && (!check || check.deleted)) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <CheckDetailsContainer
                client={client}
                check={check}
                loading={loading || aborted}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default CheckDetailsContent;
