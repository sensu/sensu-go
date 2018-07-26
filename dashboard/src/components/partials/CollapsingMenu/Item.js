import React from "react";
import PropTypes from "prop-types";
import { Context } from "./CollapsingMenu";

class Item extends React.PureComponent {
  static displayName = "CollapsingMenu.Item";

  static propTypes = {
    pinned: PropTypes.bool,
    renderMenuItem: PropTypes.func.isRequired,
    renderToolbarItem: PropTypes.func,
  };

  static defaultProps = {
    pinned: false,
    renderToolbarItem: null,
  };

  render() {
    const { pinned, renderMenuItem, renderToolbarItem } = this.props;

    return (
      <Context.Consumer>
        {({ parent, collapsed, close }) => {
          if (parent === "menu" && collapsed && !pinned) {
            return renderMenuItem({ close, collapsed });
          }

          if (parent === "buttonset" && (pinned || !collapsed)) {
            return renderToolbarItem({ collapsed });
          }

          return null;
        }}
      </Context.Consumer>
    );
  }
}

export default Item;
