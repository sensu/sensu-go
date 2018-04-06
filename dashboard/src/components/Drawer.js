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
import SettingsIcon from "material-ui-icons/Settings";
import FeedbackIcon from "material-ui-icons/Feedback";
import LogoutIcon from "material-ui-icons/ExitToApp";
import IconButton from "material-ui/IconButton";
import MenuIcon from "material-ui-icons/Menu";
import WandIcon from "../icons/Wand";
import EnvironmentIcon from "./EnvironmentIcon";
import Wordmark from "../icons/SensuWordmark";

import { logout } from "../utils/authentication";
import { makeNamespacedPath } from "./NamespaceLink";
import DrawerButton from "./DrawerButton";
import NamespaceSelector from "./NamespaceSelector";
import Preferences from "./Preferences";

const styles = theme => ({
  paper: {
    minWidth: 264,
    maxWidth: 400,
    backgroundColor: theme.palette.background.paper,
  },
  headerContainer: {
    paddingTop: "env(safe-area-inset-top)",
    backgroundColor: theme.palette.primary.dark,
  },
  header: {
    height: 172,
  },
  row: {
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
  },
  logo: {
    height: 16,
    margin: "16px 16px 0 0",
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
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object.isRequired,
    environment: PropTypes.object.isRequired,
    onToggle: PropTypes.func.isRequired,
    router: routerShape.isRequired,
    match: matchShape.isRequired,
    open: PropTypes.bool.isRequired,
  };

  state = {
    preferencesOpen: false,
  };

  handleLogout = async () => {
    await logout();
    this.props.router.push("/login");
  };

  linkTo = path => {
    const { router, match, onToggle } = this.props;
    const fullPath = makeNamespacedPath(match.params)(path);
    return () => {
      router.push(fullPath);
      onToggle();
    };
  };

  render() {
    const { viewer, environment, open, onToggle, classes } = this.props;
    const { preferencesOpen } = this.state;

    return (
      <MaterialDrawer variant="temporary" open={open} onClose={onToggle}>
        <div className={classes.paper}>
          <div className={classes.headerContainer}>
            <div className={classes.header}>
              <div className={classes.row}>
                <IconButton
                  onClick={onToggle}
                  className={classes.hamburgerButton}
                >
                  <MenuIcon />
                </IconButton>
                <Wordmark
                  alt="sensu"
                  className={classes.logo}
                  color="secondary"
                />
              </div>
              <div className={classes.row}>
                <div className={classes.namespaceIcon}>
                  <EnvironmentIcon environment={environment} size={36} />
                </div>
              </div>
              <div className={classes.row}>
                <NamespaceSelector
                  viewer={viewer}
                  className={classes.namespaceSelector}
                />
              </div>
            </div>
          </div>
          <Divider />
          <List>
            <DrawerButton
              Icon={DashboardIcon}
              primary="Dashboard"
              onClick={this.linkTo("")}
            />
            <DrawerButton
              Icon={EventIcon}
              primary="Events"
              onClick={this.linkTo("events")}
            />
            <DrawerButton
              Icon={EntityIcon}
              primary="Entities"
              onClick={this.linkTo("entities")}
            />
            <DrawerButton
              Icon={CheckIcon}
              primary="Checks"
              onClick={this.linkTo("checks")}
            />
          </List>
          <Divider />
          <List>
            <DrawerButton Icon={SettingsIcon} primary="Settings" />
            <DrawerButton
              Icon={WandIcon}
              primary="Preferences"
              onClick={() => this.setState({ preferencesOpen: true })}
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
        <Preferences
          open={preferencesOpen}
          onClose={() => this.setState({ preferencesOpen: false })}
        />
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

    fragment Drawer_environment on Environment {
      ...EnvironmentIcon_environment
    }
  `,
);

export default compose(withStyles(styles), withRouter)(DrawerContainer);
