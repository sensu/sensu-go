import React from "react";
import PropTypes from "prop-types";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";

class DrawerButton extends React.Component {
  static propTypes = {
    ...ListItem.propTypes,
    Icon: PropTypes.func.isRequired,
    primary: PropTypes.string.isRequired,
  };

  static defaultProps = {
    component: "button",
  };

  render() {
    const { Icon, primary, component, ...props } = this.props;

    return (
      <ListItem {...props} component={component}>
        <ListItemIcon>
          <Icon />
        </ListItemIcon>
        <ListItemText primary={primary} />
      </ListItem>
    );
  }
}

export default DrawerButton;
