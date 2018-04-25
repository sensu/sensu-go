import React from "react";
import PropTypes from "prop-types";

class Maybe extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    fallback: PropTypes.node,
    // eslint-disable-next-line react/require-default-props
    value: PropTypes.any,
  };

  static defaultProps = {
    // "doesn't look like anything to me"
    fallback: "nothing",
  };

  render() {
    const { children, fallback, value } = this.props;

    if (value) {
      return children(value);
    }
    return fallback;
  }
}

export default Maybe;
