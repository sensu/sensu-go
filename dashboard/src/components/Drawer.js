import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape } from "found";

import MaterialDrawer from "material-ui/Drawer";
import List from "material-ui/List";
import IconButton from "material-ui/IconButton";
import Divider from "material-ui/Divider";
import { withStyles } from "material-ui/styles";

import MenuIcon from "material-ui-icons/Menu";
import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import EventIcon from "material-ui-icons/Announcement";
import DashboardIcon from "material-ui-icons/Dashboard";
import SilenceIcon from "material-ui-icons/VolumeOff";
import HookIcon from "material-ui-icons/Link";
import HandlerIcon from "material-ui-icons/CallSplit";
import SettingsIcon from "material-ui-icons/Settings";
import FeedbackIcon from "material-ui-icons/Feedback";
import LogoutIcon from "material-ui-icons/ExitToApp";

import { logout } from "../utils/authentication";
import DrawerButton from "./DrawerButton";
import OrganizationIcon from "./OrganizationIcon";

const logo = require("../assets/logo/graphic/green.svg");

const styles = theme => ({
  paper: {
    minWidth: 264,
    maxWidth: 400,
    backgroundColor: theme.palette.background.paper,
  },
  header: {
    height: 172,
    backgroundColor: theme.palette.primary.dark,
  },
  row: {
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
  },
  // TODO create a resized version of the logo
  logo: {
    padding: "16px 0 0",
    height: 42,
    width: 24,
  },
  listItemButton: {
    padding: "16px 0 0 16px",
    color: theme.palette.primary.contrastText,
  },
  listItemContent: { padding: "0 16px 0" },
  orgIcon: { padding: "33px 0 0 16px" },
});

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
          <div className={classes.header}>
            <div className={classes.row}>
              <IconButton className={classes.listItemButton} onClick={onToggle}>
                <MenuIcon />
              </IconButton>
              <div className={classes.listItemContent}>
                <img alt="sensu" src={logo} className={classes.logo} />
              </div>
            </div>
            <div className={classes.row}>
              {/* TODO update with global variables or whatever when we get them */}
              <div className={classes.orgIcon}>
                <OrganizationIcon icon="Visibility" iconColor="#f4b2c0" />
              </div>
            </div>
          </div>
          <Divider />
          <List>
            <DrawerButton Icon={DashboardIcon} primary="Dashboard" />
            <DrawerButton Icon={EventIcon} primary="Events" href="/events" />
            <DrawerButton Icon={EntityIcon} primary="Entities" />
            <DrawerButton Icon={CheckIcon} primary="Checks" href="/checks" />
            <DrawerButton
              Icon={SilenceIcon}
              primary="Silences"
              href="/silences"
            />
            <DrawerButton Icon={HookIcon} primary="Hooks" href="/hooks" />
            <DrawerButton
              Icon={HandlerIcon}
              primary="Handlers"
              href="/handlers"
            />
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
