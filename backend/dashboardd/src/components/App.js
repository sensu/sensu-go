import React from "react";
import PropTypes from "prop-types";

import ThemeProvider from "./Theme/Provider";
import Sidebar from "./Sidebar";
import Toolbar from "./Toolbar";

class App extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
  };

  state = {
    toolbar: false,
  };

  render() {
    const toggleToolbar = () => {
      this.setState({ toolbar: !this.state.toolbar });
    };

    return (
      <ThemeProvider>
        <Toolbar toggleToolbar={toggleToolbar} />
        <Sidebar open={this.state.toolbar} />
        <div>{this.props.children}</div>
      </ThemeProvider>
    );
  }
}

export default App;
