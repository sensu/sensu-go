import React from "react";
import PropTypes from "prop-types";

import Button from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import DropdownArrow from "@material-ui/icons/ArrowDropDown";

import Menu from "@material-ui/core/Menu";
import RootRef from "@material-ui/core/RootRef";

import MenuController from "/components/controller/MenuController";

class ButtonMenu extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    label: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
  };

  render() {
    const { label, children } = this.props;

    return (
      <MenuController
        renderMenu={({ anchorEl, close }) => (
          <Menu open onClose={close} anchorEl={anchorEl}>
            {React.Children.map(children, child => {
              const onClick = event => {
                if (child.props.onClick) {
                  child.props.onClick(event);
                  if (event.defaultPrevented) {
                    return;
                  }
                }

                if (child.props.value !== undefined) {
                  this.props.onChange(child.props.value);
                }

                close();
              };

              return React.cloneElement(child, { onClick });
            })}
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
