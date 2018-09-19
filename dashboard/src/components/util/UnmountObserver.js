import React from "react";
import PropTypes from "prop-types";

class UnmountObserver extends React.PureComponent {
  static propTypes = {
    onUnmount: PropTypes.func.isRequired,
  };

  componentWillUnmount() {
    this.props.onUnmount();
  }

  render() {
    return null;
  }
}

export default UnmountObserver;
