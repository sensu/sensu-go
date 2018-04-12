import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";

import RestrictUnauthenticated from "./RestrictUnauthenticated";
import AppFrame from "./AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    children: PropTypes.element,
    match: PropTypes.object.isRequired,
  };

  static defaultProps = { children: null };

  static query = gql`
    query AppWrapperQuery($environment: String!, $organization: String!) {
      viewer {
        ...AppFrame_viewer
      }
      environment(organization: $organization, environment: $environment) {
        ...AppFrame_environment
      }
    }
    ${AppFrame.fragments.viewer}
    ${AppFrame.fragments.environment}
  `;

  render() {
    const { match, children } = this.props;

    return (
      <RestrictUnauthenticated>
        <Query query={AppWrapper.query} variables={match.params}>
          {({ data: { viewer, environment } = {}, loading, error }) => {
            // TODO: Connect this error handler to display a blocking error
            // alert
            if (error) throw error;

            return (
              <AppFrame
                loaded={!loading}
                viewer={viewer}
                environment={environment}
              >
                {children}
              </AppFrame>
            );
          }}
        </Query>
      </RestrictUnauthenticated>
    );
  }
}

export default AppWrapper;
