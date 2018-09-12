import React from "react";
import PropTypes from "prop-types";

class Item extends React.PureComponent {
  static displayName = "ToolbarMenu.Item";

  static propTypes = {
    children: PropTypes.oneOfType([PropTypes.node, PropTypes.func]).isRequired,
    visible: PropTypes.oneOf(["if-room", "always", "never"]),
  };

  static defaultProps = {
    visible: "if-room",
  };

  render() {
    const { children: childrenProp, visible } = this.props;

    if (visible === "hidden") {
      return null;
    }

    let children = childrenProp;
    if (typeof childrenProp === "function") {
      children = childrenProp();
    }

    return children;
  }
}

export default Item;
