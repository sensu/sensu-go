import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape } from "found";

import MaterialDrawer from "material-ui/Drawer";
import List, { ListItem } from "material-ui/List";
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
import LogoutIcon from "material-ui-icons/ExitToApp";

import { logout } from "../utils/authentication";
import DrawerButton from "./DrawerButton";

const logo = require("../assets/logo/full/silver-green.svg");

const styles = theme => {
  const listItemStyles = listItemIconStyles(theme);

  return {
    paper: {
      minWidth: 264,
      maxWidth: 400,
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
    router: routerShape.isRequired,
    open: PropTypes.bool.isRequired,
  };

  handleLogout = async () => {
    await logout();
    this.props.router.push("/login");
  };

  render() {
    const { open, onToggle, classes } = this.props;

    return (
      <MaterialDrawer type="temporary" open={open} onClose={onToggle}>
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
            <DrawerButton Icon={DashboardIcon} primary="Dashboard" />
            <DrawerButton Icon={EventIcon} primary="Events" href="/events" />
            <DrawerButton Icon={EntityIcon} primary="Entities" />
            <DrawerButton Icon={CheckIcon} primary="Checks" href="/checks" />
          </List>
          <Divider />
          <List>
            <DrawerButton Icon={SettingsIcon} primary="Settings" />
            <DrawerButton
              Icon={FeedbackIcon}
              primary="Feedback"
              href="https://www.sensuapp.org"
            />
            <DrawerButton
              Icon={LogoutIcon}
              primary="Sign out"
              onClick={this.handleLogout}
            />
          </List>
        </div>
      </MaterialDrawer>
    );
  }
}

export default compose(withStyles(styles), withRouter)(Drawer);
