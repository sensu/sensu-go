import React from "react";
import PropTypes from "prop-types";
import createBroadcast from "brcast";
import ThemeProvider from "./ThemeProvider";

class AppThemeProvider extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
  };

  static childContextTypes = {
    CURRENT_THEME: PropTypes.object.isRequired,
  };

  constructor(props, context) {
    super(props, context);
    this.broadcast.setState({
      theme: "sensu",
      dark: false,
    });
  }

  state = {};

  getChildContext() {
    return { CURRENT_THEME: this.broadcast };
  }

  componentDidMount() {
    this.unsubscribeId = this.broadcast.subscribe(st => this.setState(st));
  }

  componentWillUnmount() {
    if (this.unsubscribeId !== null) {
      this.broadcast.unsubscribe(this.unsubscribeId);
    }
  }

  broadcast = createBroadcast();
  unsubscribeId = null;

  render() {
    const { theme, dark } = this.state;
    const { children } = this.props;
    return (
      <ThemeProvider theme={theme} dark={dark}>
        {children}
      </ThemeProvider>
    );
  }
}

export default AppThemeProvider;
