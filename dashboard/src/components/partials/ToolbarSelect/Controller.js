import React from "react";
import PropTypes from "prop-types";

import MenuController from "/components/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";

import Menu from "./Menu";

class Controller extends React.PureComponent {
  static displayName = "ToolbarSelect.Controller";

  static propTypes = {
    children: PropTypes.func.isRequired,
    options: PropTypes.arrayOf(PropTypes.node).isRequired,
    onChange: PropTypes.func.isRequired,
    onClose: PropTypes.func,
  };

  static defaultProps = {
    onClose: () => null,
  };

  render() {
    const { children, onChange, onClose, options } = this.props;

    return (
      <MenuController
        renderMenu={({ anchorEl, close }) => (
          <Menu
            anchorEl={anchorEl}
            onChange={onChange}
            onClose={() => {
              onClose();
              close();
            }}
          >
            {options}
          </Menu>
        )}
      >
        {({ open, ref }) => (
          <RootRef rootRef={ref}>{children({ open })}</RootRef>
        )}
      </MenuController>
    );
  }
}

export default Controller;
