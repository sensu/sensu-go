import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";

import DefaultThemeProvider from "./Theme/Provider";
import RestrictUnauthenticated from "./RestrictUnauthenticated";
import AppFrame from "./AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  static query = graphql`
    query AppWrapperQuery {
      viewer {
        ...AppFrame_viewer
      }
    }
  `;

  render() {
    const { viewer, children } = this.props;
    return (
      <RestrictUnauthenticated>
        <DefaultThemeProvider>
          <AppFrame viewer={viewer}>{children}</AppFrame>
        </DefaultThemeProvider>
      </RestrictUnauthenticated>
    );
  }
}

export default AppWrapper;
