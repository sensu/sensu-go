import React from "react";
import PropTypes from "prop-types";
import { withRouter, routerShape } from "found";
import { styles as listItemIconStyles } from "material-ui/List/ListItemIcon";
import { withStyles } from "material-ui/styles";
import List, { ListItem, ListItemIcon, ListItemText } from "material-ui/List";

const styles = theme => {
  const listItemStyles = listItemIconStyles(theme);

  return {
    listItem: listItemStyles.root,
    listItemIcon: {
      padding: "0 0",
    },
  };
};

class QuickNavButton extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
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
    const { classes, Icon, router, primary, onClick, ...props } = this.props;
    const handleClick = () => this.props.router.push(this.props.href);

    return (
      <List>
        <ListItem button onClick={onClick || handleClick}>
          <ListItemIcon className={classes.ListItemIcon}>
            <Icon />
          </ListItemIcon>
        </ListItem>
        <ListItem button onClick={onClick || handleClick}>
          <ListItemText primary={primary} {...props} />
        </ListItem>
      </List>
    );
  }
}

export default withRouter(withStyles(styles)(QuickNavButton));
