import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape, matchShape } from "found";

import MaterialDrawer from "material-ui/Drawer";
import List from "material-ui/List";
import Divider from "material-ui/Divider";
import { withStyles } from "material-ui/styles";

import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import EventIcon from "material-ui-icons/Notifications";
import DashboardIcon from "material-ui-icons/Dashboard";
import SilenceIcon from "material-ui-icons/VolumeOff";
import HookIcon from "material-ui-icons/Link";
import HandlerIcon from "material-ui-icons/CallSplit";
import SettingsIcon from "material-ui-icons/Settings";
import FeedbackIcon from "material-ui-icons/Feedback";
import LogoutIcon from "material-ui-icons/ExitToApp";
import WandIcon from "../icons/Wand";

import { logout } from "../utils/authentication";
import { makeNamespacedPath } from "./NamespaceLink";
import DrawerHeader from "./DrawerHeader";
import DrawerButton from "./DrawerButton";

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
});

class Drawer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    onToggle: PropTypes.func.isRequired,
    router: routerShape.isRequired,
    match: matchShape.isRequired,
    open: PropTypes.bool.isRequired,
  };

  handleLogout = async () => {
    await logout();
    this.props.router.push("/login");
  };

  render() {
    const { open, router, match, onToggle, classes } = this.props;
    const linkTo = path => {
      const fullPath = makeNamespacedPath(match.params)(path);
      return () => {
        router.push(fullPath);
        onToggle();
      };
    };

    return (
      <MaterialDrawer type="temporary" open={open} onClose={onToggle}>
        <div className={classes.paper}>
          <DrawerHeader onToggle={onToggle} />
          <Divider />
          <List>
            <DrawerButton Icon={DashboardIcon} primary="Dashboard" />
            <DrawerButton
              Icon={EventIcon}
              primary="Events"
              onClick={linkTo("events")}
            />
            <DrawerButton
              Icon={EntityIcon}
              primary="Entities"
              onClick={linkTo("entities")}
            />
            <DrawerButton
              Icon={CheckIcon}
              primary="Checks"
              onClick={linkTo("checks")}
            />
            <DrawerButton
              Icon={SilenceIcon}
              primary="Silences"
              onClick={linkTo("silences")}
            />
            <DrawerButton Icon={HookIcon} primary="Hooks" href="hooks" />
            <DrawerButton
              Icon={HandlerIcon}
              primary="Handlers"
              onClick={linkTo("handlers")}
            />
          </List>
          <Divider />
          <List>
            <DrawerButton Icon={SettingsIcon} primary="Settings" />
            <DrawerButton
              Icon={WandIcon}
              primary="Preferences"
              onClick={linkTo("preferences")}
            />
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
