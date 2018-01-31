import React from "react";
import PropTypes from "prop-types";

import DefaultThemeProvider from "./Theme/Provider";
import RestrictUnauthenticated from "./RestrictUnauthenticated";
import AppFrame from "./AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  render() {
    const { children } = this.props;
    return (
      <RestrictUnauthenticated>
        <DefaultThemeProvider>
          <AppFrame>{children}</AppFrame>
        </DefaultThemeProvider>
      </RestrictUnauthenticated>
    );
  }
}

export default AppWrapper;
