import React from "react";
import PropTypes from "prop-types";

import DefaultThemeProvider from "./Theme/Provider";
import AppFrame from "./AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
  };

  render() {
    return (
      <DefaultThemeProvider>
        <AppFrame>{this.props.children}</AppFrame>
      </DefaultThemeProvider>
    );
  }
}

export default AppWrapper;
