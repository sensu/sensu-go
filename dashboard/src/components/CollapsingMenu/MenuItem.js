import React from "react";
import PropTypes from "prop-types";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import Button from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import BaseItem from "./MenuItemBase";

class CollapsingMenuItem extends React.PureComponent {
  static propTypes = {
    icon: PropTypes.node,
    title: PropTypes.string.isRequired,
    subtitle: PropTypes.string,
    onClick: PropTypes.func.isRequired,
  };

  static defaultProps = {
    icon: null,
    subtitle: null,
  };

  render() {
    const { icon, title, subtitle, onClick, ...props } = this.props;

    return (
      <BaseItem
        renderMenuItem={
          <MenuItem onClick={onClick}>
            <ListItemIcon>{icon}</ListItemIcon>
            <ListItemText inset primary={title} secondary={subtitle} />
          </MenuItem>
        }
        renderButton={
          <Button onClick={onClick}>
            <ButtonIcon>{icon}</ButtonIcon>
            {title}
          </Button>
        }
        {...props}
      />
    );
  }
}

export default CollapsingMenuItem;
