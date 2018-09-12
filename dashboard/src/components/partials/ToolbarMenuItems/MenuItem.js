import React from "react";
import PropTypes from "prop-types";

import { Context } from "/components/partials/ToolbarMenu";

import AdaptiveMenuItem from "./AdaptiveMenuItem";

class MenuItem extends React.Component {
  static displayName = "ToolbarMenuItems.MenuItem";

  static propTypes = {
    autoClose: PropTypes.bool,
    onClick: PropTypes.func,
  };

  static defaultProps = {
    autoClose: true,
    onClick: () => null,
  };

  render() {
    const { autoClose, onClick: onClickProp, ...props } = this.props;

    return (
      <Context.Consumer>
        {({ collapsed, close }) => {
          const onClick = ev => {
            onClickProp(ev);
            if (autoClose) {
              close(ev);
            }
          };

          return (
            <AdaptiveMenuItem
              collapsed={collapsed}
              onClick={onClick}
              {...props}
            />
          );
        }}
      </Context.Consumer>
    );
  }
}

export default MenuItem;
