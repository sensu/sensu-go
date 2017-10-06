import React from "react";
import PropTypes from "prop-types";
import { MuiThemeProvider } from "material-ui/styles";

import Default from "./Default";

class ThemeProvider extends React.Component {
  static propTypes = {
    theme: PropTypes.object, // eslint-disable-line react/forbid-prop-types
    children: PropTypes.element.isRequired,
  };

  static defaultProps = {
    theme: Default,
  };

  render() {
    return (
      <MuiThemeProvider theme={this.props.theme}>
        {this.props.children}
      </MuiThemeProvider>
    );
  }
}

export default ThemeProvider;
