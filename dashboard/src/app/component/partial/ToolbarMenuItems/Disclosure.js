import React from "react";
import PropTypes from "prop-types";

import ArrowDown from "@material-ui/icons/ArrowDropDown";
import ArrowRight from "@material-ui/icons/KeyboardArrowRight";
import MenuItem from "./AdaptiveMenuItem";

const buttonIcon = <ArrowDown />;
const menuIcon = <ArrowRight />;

class Disclosure extends React.Component {
  static displayName = "ToolbarMenuItems.Disclosure";

  static propTypes = {
    collapsed: PropTypes.bool,
  };

  static defaultProps = {
    collapsed: false,
  };

  render() {
    const { collapsed, ...props } = this.props;

    return (
      <MenuItem
        collapsed={collapsed}
        ornament={collapsed ? menuIcon : buttonIcon}
        {...props}
      />
    );
  }
}

export default Disclosure;
