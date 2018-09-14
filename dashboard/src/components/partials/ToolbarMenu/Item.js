import React from "react";
import PropTypes from "prop-types";
import ResizeObserver from "react-resize-observer";

import { Context } from "./Menu";

class Item extends React.PureComponent {
  static displayName = "ToolbarMenu.Item";

  static propTypes = {
    children: PropTypes.oneOfType([PropTypes.node, PropTypes.func]).isRequired,
    id: PropTypes.string.isRequired, // eslint-disable-line react/no-unused-prop-types
    onResize: PropTypes.func,
    visible: PropTypes.oneOf(["if-room", "always", "never"]), // eslint-disable-line react/no-unused-prop-types
  };

  static defaultProps = {
    onResize: null,
    visible: "if-room",
  };

  render() {
    const { children: childrenProp } = this.props;

    let children = childrenProp;
    if (typeof children === "function") {
      children = <Context.Consumer>{childrenProp}</Context.Consumer>;
    }

    if (this.props.onResize) {
      return (
        <div style={{ position: "relative", display: "inline" }}>
          <ResizeObserver onResize={this.props.onResize} />
          {children}
        </div>
      );
    }

    return children;
  }
}

export default Item;
