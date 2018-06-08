import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Query from "/components/util/Query";
import Loader from "/components/util/Loader";

import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import EntityDetailsContainer from "/components/partials/EntityDetailsContainer";

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
        variables={{ ...params, ns }}
      >
        {({ data: { entity } = {}, loading, aborted }) => {
          if (!loading && !aborted && (!entity || entity.deleted)) {
            return <NotFoundView />;
          }

          return (
            <AppContent>
              <Loader loading={loading || aborted} passthrough>
                {entity && <EntityDetailsContainer entity={entity} />}
              </Loader>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EntityDetailsContent;
