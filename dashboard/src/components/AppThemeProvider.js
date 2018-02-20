import React from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import ThemeProvider from "./ThemeProvider";

class AppThemeProvider extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
    theme: PropTypes.shape({
      theme: PropTypes.string,
      dark: PropTypes.bool,
    }).isRequired,
  };

  render() {
    const { theme, children } = this.props;
    return (
      <ThemeProvider theme={theme.theme} dark={theme.dark}>
        {children}
      </ThemeProvider>
    );
  }
}

const withTheme = connect(state => ({ theme: state.theme }));
export default withTheme(AppThemeProvider);
