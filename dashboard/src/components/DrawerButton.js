import React from "react";
import PropTypes from "prop-types";
import { ListItem, ListItemIcon, ListItemText } from "material-ui/List";

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
