import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";

import RestrictUnauthenticated from "./RestrictUnauthenticated";
import AppFrame from "./AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
    environment: PropTypes.objectOf(PropTypes.any).isRequired,
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  static query = graphql`
    query AppWrapperQuery($environment: String!, $organization: String!) {
      viewer {
        ...AppFrame_viewer
      }
      environment(organization: $organization, environment: $environment) {
        ...AppFrame_environment
      }
    }
  `;

  render() {
    const { viewer, environment, children } = this.props;
    return (
      <RestrictUnauthenticated>
        <AppFrame viewer={viewer} environment={environment}>
          {children}
        </AppFrame>
      </RestrictUnauthenticated>
    );
  }
}

export default AppWrapper;
