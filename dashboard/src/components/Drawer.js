import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape, matchShape } from "found";
import { createFragmentContainer, graphql } from "react-relay";

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
import IconButton from "material-ui/IconButton";
import MenuIcon from "material-ui-icons/Menu";
import WandIcon from "../icons/Wand";
import OrganizationIcon from "./OrganizationIcon";

import { logout } from "../utils/authentication";
import { makeNamespacedPath } from "./NamespaceLink";
import DrawerButton from "./DrawerButton";
import NamespaceSelector from "./NamespaceSelector";
import logo from "../assets/logo/wordmark/green.svg";

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
  logo: {
    height: 24,
    margin: "12px 12px 0 0",
  },
  namespaceSelector: {
    margin: "8px 0 -8px 0",
    width: "100%",
  },
  namespaceIcon: {
    margin: "24px 0 0 16px",
  },
  hamburgerButton: {
    color: theme.palette.primary.contrastText,
  },
});

class Drawer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    // eslint-disable-next-line react/forbid-prop-types
    viewer: PropTypes.object.isRequired,
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
    const { viewer, open, router, match, onToggle, classes } = this.props;
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
          <div className={classes.header}>
            <div className={classes.row}>
              <IconButton
                onClick={onToggle}
                className={classes.hamburgerButton}
              >
                <MenuIcon />
              </IconButton>
              <img alt="sensu" src={logo} className={classes.logo} />
            </div>
            <div className={classes.row}>
              {/* TODO update with global variables or whatever when we get them */}
              <div className={classes.namespaceIcon}>
                <OrganizationIcon
                  icon="HalfHeart"
                  iconColor="#FA8072"
                  size={36}
                />
              </div>
            </div>
            <div className={classes.row}>
              <NamespaceSelector
                viewer={viewer}
                className={classes.namespaceSelector}
              />
            </div>
          </div>
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

const DrawerContainer = createFragmentContainer(
  Drawer,
  graphql`
    fragment Drawer_viewer on Viewer {
      ...NamespaceSelector_viewer
    }
  `,
);

export default compose(withStyles(styles), withRouter)(DrawerContainer);
