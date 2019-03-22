import React from "react";
import PropTypes from "prop-types";

class ConfirmAction extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  state = {
    isOpen: false,
  };

  render() {
    const { isOpen } = this.state;
    const childArgs = {
      isOpen,
      open: () => this.setState({ isOpen: true }),
      close: () => this.setState({ isOpen: false }),
    };
    return this.props.children(childArgs);
  }
}

export default ConfirmAction;
