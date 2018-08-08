import React from "react";
import PropTypes from "prop-types";

import Button from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import DropdownArrow from "@material-ui/icons/ArrowDropDown";
import MenuController from "/components/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";

import Menu from "./Menu";

class ButtonMenu extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    label: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
  };

  render() {
    const { label, children, onChange } = this.props;

    return (
      <MenuController
        renderMenu={({ anchorEl, close }) => (
          <Menu anchorEl={anchorEl} onChange={onChange} onClose={close}>
            {children}
          </Menu>
        )}
      >
        {({ open, ref }) => (
          <RootRef rootRef={ref}>
            <Button onClick={open}>
              {label}
              <ButtonIcon alignment="right">
                <DropdownArrow />
              </ButtonIcon>
            </Button>
          </RootRef>
        )}
      </MenuController>
    );
  }
}

export default ButtonMenu;
