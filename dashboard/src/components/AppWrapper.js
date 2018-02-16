import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";

import ThemeProvider from "./AppThemeProvider";
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
        <ThemeProvider>
          <AppFrame viewer={viewer}>{children}</AppFrame>
        </ThemeProvider>
      </RestrictUnauthenticated>
    );
  }
}

export default AppWrapper;
