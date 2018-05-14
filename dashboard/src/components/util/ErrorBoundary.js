import React from "react";
import PropTypes from "prop-types";

import ReactError from "/errors/ReactError";

class ErrorBoundary extends React.PureComponent {
  static propTypes = {
    handle: PropTypes.func.isRequired,
  };

  componentDidCatch(error, info) {
    if (this.props.handle) {
      this.props.handle(new ReactError(error, info));
    } else {
      throw error;
    }
  }

  render() {
    // eslint-disable-next-line react/prop-types
    return this.props.children;
  }
}

export default ErrorBoundary;
