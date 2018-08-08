import React from "react";
import PropTypes from "prop-types";

import Button from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import DropdownArrow from "@material-ui/icons/ArrowDropDown";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import MenuController from "/components/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";
import SubdirectoryArrowLeft from "@material-ui/icons/KeyboardArrowLeft";

import Item from "./Item";

class SubMenu extends React.Component {
  static displayName = "CollapsingMenu.SubMenu";

  static propTypes = {
    disabled: PropTypes.bool,
    color: PropTypes.string,
    title: PropTypes.string.isRequired,
    subtitle: PropTypes.string,
    pinned: PropTypes.bool,
    renderMenu: PropTypes.func.isRequired,
  };

  static defaultProps = {
    disabled: false,
    color: "inherit",
    subtitle: null,
    pinned: false,
  };

  render() {
    const { title, subtitle, pinned, renderMenu, ...props } = this.props;

    return (
      <Item
        pinned={pinned}
        renderMenuItem={({ close: closeRenderProp }) => {
          const renderMenuProp = ({ anchorEl, close: closeProp }) => {
            let close = closeRenderProp;
            if (closeProp) {
              close = () => {
                closeProp();
                closeRenderProp();
              };
            }

            return renderMenu({ anchorEl, close });
          };

          return (
            <MenuController renderMenu={renderMenuProp}>
              {({ open, ref }) => (
                <RootRef rootRef={ref}>
                  <MenuItem onClick={open}>
                    <ListItemIcon>
                      <SubdirectoryArrowLeft />
                    </ListItemIcon>
                    <ListItemText inset primary={title} secondary={subtitle} />
                  </MenuItem>
                </RootRef>
              )}
            </MenuController>
          );
        }}
        renderToolbarItem={() => {
          const { color, disabled } = props;

          return (
            <MenuController renderMenu={renderMenu}>
              {({ open, ref }) => (
                <RootRef rootRef={ref}>
                  <Button color={color} disabled={disabled} onClick={open}>
                    {title}
                    <ButtonIcon alignment="right">
                      <DropdownArrow />
                    </ButtonIcon>
                  </Button>
                </RootRef>
              )}
            </MenuController>
          );
        }}
      />
    );
  }
}

export default SubMenu;
