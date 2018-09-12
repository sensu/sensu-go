import React from "react";
import PropTypes from "prop-types";

import ArrowDown from "@material-ui/icons/ArrowDropDown";
import ArrowRight from "@material-ui/icons/KeyboardArrowRight";
import { Context } from "/components/partials/ToolbarMenu";
import MenuController from "/components/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";

import MenuItem from "./AdaptiveMenuItem";

class Submenu extends React.Component {
  static displayName = "ToolbarMenuItems.Submenu";

  static propTypes = {
    autoClose: PropTypes.bool,
    renderMenu: PropTypes.func.isRequired,
    ...MenuItem.propTypes,
  };

  static defaultProps = {
    autoClose: true,
    renderMenu: () => null,
  };

  render() {
    const { autoClose, renderMenu, ...props } = this.props;

    return (
      <Context.Consumer>
        {({ collapsed, close: closeParent }) => (
          <MenuController
            renderMenu={({ close: closeMenu, ...renderProps }) => {
              let close = closeMenu;
              if (autoClose) {
                close = () => {
                  closeMenu();
                  closeParent();
                };
              }

              return renderMenu({ ...renderProps, close, closeParent });
            }}
          >
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <MenuItem
                  {...props}
                  collapsed={collapsed}
                  onClick={open}
                  ornament={collapsed ? <ArrowRight /> : <ArrowDown />}
                />
              </RootRef>
            )}
          </MenuController>
        )}
      </Context.Consumer>
    );
  }
}

export default Submenu;
