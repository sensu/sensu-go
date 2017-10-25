import React from "react";
import PropTypes from "prop-types";
import { withRouter, routerShape } from "found";
import { ListItem, ListItemIcon, ListItemText } from "material-ui/List";

class DrawerButton extends React.Component {
  static propTypes = {
    Icon: PropTypes.func.isRequired,
    primary: PropTypes.string.isRequired,
    router: routerShape.isRequired,
    href: PropTypes.string,
    onClick: PropTypes.func,
  };

  static defaultProps = {
    onClick: null,
    href: "",
  };

  render() {
    const { Icon, router, primary, onClick, ...props } = this.props;
    const handleClick = () => this.props.router.push(this.props.href);

    return (
      <ListItem button onClick={onClick || handleClick}>
        <ListItemIcon>
          <Icon />
        </ListItemIcon>
        <ListItemText primary={primary} {...props} />
      </ListItem>
    );
  }
}

export default withRouter(DrawerButton);
