import React from "react";
import PropTypes from "prop-types";
import { MuiThemeProvider } from "material-ui/styles";
import * as themes from "../themes";

class ThemeProvider extends React.Component {
  static propTypes = {
    theme: PropTypes.string,
    dark: PropTypes.bool,
    children: PropTypes.element.isRequired,
  };

  static defaultProps = {
    theme: "sensu",
    dark: false,
  };

  componentWillMount() {
    const type = this.props.dark ? "dark" : "light";
    const theme = themes[this.props.theme];
    this.theme = theme(type);
  }

  componentWillReceiveProps(nextProps) {
    if (
      nextProps.theme !== this.props.theme ||
      nextProps.dark !== this.props.dark
    ) {
      const type = nextProps.dark ? "dark" : "light";
      const theme = themes[nextProps.theme];
      this.theme = theme(type);
    }
  }

  render() {
    return (
      <MuiThemeProvider theme={this.theme}>
        {this.props.children}
      </MuiThemeProvider>
    );
  }
}

export default ThemeProvider;
