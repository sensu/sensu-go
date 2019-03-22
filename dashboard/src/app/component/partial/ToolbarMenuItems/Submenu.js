import React from "react";
import PropTypes from "prop-types";

import { Context } from "/app/component/partial/ToolbarMenu";
import MenuController from "/lib/component/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";

import Disclosure from "./Disclosure";

class Submenu extends React.Component {
  static displayName = "ToolbarMenuItems.Submenu";

  static propTypes = {
    autoClose: PropTypes.bool,
    renderMenu: PropTypes.func,
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
                <Disclosure collapsed={collapsed} onClick={open} {...props} />
              </RootRef>
            )}
          </MenuController>
        )}
      </Context.Consumer>
    );
  }
}

export default Submenu;
