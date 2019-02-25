import React from "react";
import PropTypes from "prop-types";

import Icon from "./SensuIcon";
import Wordmark from "./SensuWordmark";

class SensuLogo extends React.PureComponent {
  static propTypes = {
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  };

  static defaultProps = {
    component: "span",
  };

  render() {
    const { component: Component, ...props } = this.props;

    return (
      <Component {...props}>
        <Icon />
        <Wordmark
          style={{
            transform: "scale(0.55)",
            transformOrigin: "left bottom",
          }}
        />
      </Component>
    );
  }
}

export default SensuLogo;
