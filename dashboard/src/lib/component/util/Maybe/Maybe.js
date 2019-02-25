import React from "react";
import PropTypes from "prop-types";
import isObjectLike from "lodash/isObjectLike";
import isArrayLike from "lodash/isArrayLike";

function isTruthy(val) {
  if (isObjectLike(val)) {
    return Object.keys(val).length > 0;
  }
  if (isArrayLike(val)) {
    return val.length > 0;
  }
  return !!val;
}

// Maybe can be used to conditionally render a given value if it is truthy.
// If, however, the value is falsey the configured fallback is returned or an
// empty fragement.
class Maybe extends React.Component {
  static propTypes = {
    children: PropTypes.func,
    fallback: PropTypes.node,
    // eslint-disable-next-line react/require-default-props
    value: PropTypes.any,
  };

  static defaultProps = {
    // "doesn't look like anything to me"
    fallback: <React.Fragment />,
    children: null,
  };

  render() {
    const { children, fallback, value } = this.props;

    if (isTruthy(value)) {
      if (typeof children === "function") {
        return children ? children(value) : value;
      } else if (children) {
        return children;
      }
      return value;
    }
    return fallback;
  }
}

export default Maybe;
