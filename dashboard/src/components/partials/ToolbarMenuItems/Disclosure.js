import React from "react";

import ArrowDown from "@material-ui/icons/ArrowDropDown";
import ArrowRight from "@material-ui/icons/KeyboardArrowRight";
import { Context } from "/components/partials/ToolbarMenu";

import MenuItem from "./AdaptiveMenuItem";

const buttonIcon = <ArrowDown />;
const menuIcon = <ArrowRight />;

class Disclosure extends React.Component {
  static displayName = "ToolbarMenuItems.Disclosure";

  render() {
    return (
      <Context.Consumer>
        {({ collapsed }) => (
          <MenuItem
            {...this.props}
            collapsed={collapsed}
            ornament={collapsed ? menuIcon : buttonIcon}
          />
        )}
      </Context.Consumer>
    );
  }
}

export default Disclosure;
