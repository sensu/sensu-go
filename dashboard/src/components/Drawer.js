import React from "react";
import PropTypes from "prop-types";

import MaterialDrawer from "material-ui/Drawer";
import List, { ListItem, ListItemIcon, ListItemText } from "material-ui/List";
import { styles as listItemIconStyles } from "material-ui/List/ListItemIcon";
import IconButton from "material-ui/IconButton";
import Divider from "material-ui/Divider";
import { withStyles } from "material-ui/styles";

import MenuIcon from "material-ui-icons/Menu";
import EntityIcon from "material-ui-icons/DevicesOther";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import EventIcon from "material-ui-icons/Announcement";
import DashboardIcon from "material-ui-icons/Dashboard";
import SettingsIcon from "material-ui-icons/Settings";
import FeedbackIcon from "material-ui-icons/Feedback";

const logo = require("../assets/logo.png");

const styles = theme => {
  const listItemStyles = listItemIconStyles(theme);

  return {
    paper: {
      width: 280,
      backgroundColor: theme.palette.background.paper,
    },
    logo: { height: "inherit" },
    listItemButton: listItemStyles.root, // TODO ...
    listItemContent: {
      flex: "1 1 auto",
      padding: "0 16px",
      height: listItemStyles.root.height,
      "&:first-child": {
        paddingLeft: theme.spacing.unit * 7,
      },
    },
  };
};

class Drawer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    onToggle: PropTypes.func.isRequired,
    open: PropTypes.bool.isRequired,
  };

  render() {
    const { open, onToggle, classes } = this.props;

    return (
      <MaterialDrawer type="temporary" open={open} onRequestClose={onToggle}>
        <div className={classes.paper}>
          <List>
            <ListItem>
              <IconButton className={classes.listItemButton} onClick={onToggle}>
                <MenuIcon />
              </IconButton>
              <div className={classes.listItemContent}>
                <img alt="sensu" src={logo} className={classes.logo} />
              </div>
            </ListItem>
          </List>
          <Divider />
          <List>
            <ListItem button>
              <ListItemIcon>
                <DashboardIcon />
              </ListItemIcon>
              <ListItemText primary="Dashboard" />
            </ListItem>
            <ListItem button selected>
              <ListItemIcon>
                <EventIcon />
              </ListItemIcon>
              <ListItemText primary="Events" />
            </ListItem>
            <ListItem button>
              <ListItemIcon>
                <EntityIcon />
              </ListItemIcon>
              <ListItemText primary="Entities" />
            </ListItem>
            <ListItem button>
              <ListItemIcon>
                <CheckIcon />
              </ListItemIcon>
              <ListItemText primary="Checks" />
            </ListItem>
          </List>
          <Divider />
          <List>
            <ListItem button>
              <ListItemIcon>
                <SettingsIcon />
              </ListItemIcon>
              <ListItemText primary="Settings" />
            </ListItem>
            <ListItem button>
              <ListItemIcon>
                <FeedbackIcon />
              </ListItemIcon>
              <ListItemText primary="Feedback" />
            </ListItem>
          </List>
        </div>
      </MaterialDrawer>
    );
  }
}

export default withStyles(styles)(Drawer);
