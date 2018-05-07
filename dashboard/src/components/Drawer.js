import React from "react";
import PropTypes from "prop-types";
import { Route, Link } from "react-router-dom";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import { compose } from "recompose";

import MaterialDrawer from "material-ui/Drawer";
import List from "material-ui/List";
import Divider from "material-ui/Divider";
import { withStyles } from "material-ui/styles";

import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import EventIcon from "material-ui-icons/Notifications";
import FeedbackIcon from "material-ui-icons/Feedback";
import LogoutIcon from "material-ui-icons/ExitToApp";
import IconButton from "material-ui/IconButton";
import MenuIcon from "material-ui-icons/Menu";

import WandIcon from "/icons/Wand";
import Wordmark from "/icons/SensuWordmark";

import EnvironmentIcon from "/components/EnvironmentIcon";
import DrawerButton from "/components/DrawerButton";
import NamespaceSelector from "/components/NamespaceSelector";
import Preferences from "/components/Preferences";
import Loader from "/components/Loader";

import invalidateTokens from "/mutations/invalidateTokens";

const linkPath = (params, path) => {
  const { organization, environment } = params;
  return `/${organization}/${environment}/${path}`;
};

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

  drawer: {
    display: "block",
  },
});

class Drawer extends React.Component {
  static propTypes = {
    client: PropTypes.object.isRequired,
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object,
    environment: PropTypes.object,
    onToggle: PropTypes.func.isRequired,
    open: PropTypes.bool.isRequired,
    loading: PropTypes.bool,
  };

  static defaultProps = { loading: false, viewer: null, environment: null };

  static fragments = {
    viewer: gql`
      fragment Drawer_viewer on Viewer {
        ...NamespaceSelector_viewer
      }

      ${NamespaceSelector.fragments.viewer}
      ${NamespaceSelector.fragments.environment}
    `,

    environment: gql`
      fragment Drawer_environment on Environment {
        ...EnvironmentIcon_environment
        ...NamespaceSelector_environment
      }

      ${NamespaceSelector.fragments.environment}
      ${EnvironmentIcon.fragments.environment}
    `,
  };

  state = {
    preferencesOpen: false,
  };

  render() {
    const {
      client,
      loading,
      viewer,
      environment,
      open,
      onToggle,
      classes,
    } = this.props;
    const { preferencesOpen } = this.state;

    return (
      <MaterialDrawer
        variant="temporary"
        className={classes.drawer}
        open={open}
        onClose={onToggle}
      >
        <Loader passhrough loading={loading}>
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
                    {environment && (
                      <EnvironmentIcon environment={environment} size={36} />
                    )}
                  </div>
                </div>
                <div className={classes.row}>
                  <NamespaceSelector
                    viewer={viewer}
                    environment={environment}
                    className={classes.namespaceSelector}
                  />
                </div>
              </div>
            </div>
            <Divider />
            <Route
              path="/:organization/:environment"
              render={({ match: { params } }) => (
                <List>
                  <DrawerButton
                    Icon={EventIcon}
                    primary="Events"
                    component={Link}
                    onClick={onToggle}
                    to={linkPath(params, "events")}
                  />
                  <DrawerButton
                    Icon={EntityIcon}
                    primary="Entities"
                    component={Link}
                    onClick={onToggle}
                    to={linkPath(params, "entities")}
                  />
                  <DrawerButton
                    Icon={CheckIcon}
                    primary="Checks"
                    component={Link}
                    onClick={onToggle}
                    to={linkPath(params, "checks")}
                  />
                </List>
              )}
            />
            <Divider />
            <List>
              <DrawerButton
                Icon={WandIcon}
                primary="Preferences"
                onClick={() => {
                  this.setState({ preferencesOpen: true });
                }}
              />
              <DrawerButton
                Icon={FeedbackIcon}
                primary="Feedback"
                component="a"
                href="https://github.com/sensu/sensu-go/issues"
              />
              <DrawerButton
                Icon={LogoutIcon}
                primary="Sign out"
                onClick={() => {
                  onToggle();
                  invalidateTokens(client);
                }}
              />
            </List>
          </div>
        </Loader>
        <Preferences
          open={preferencesOpen}
          onClose={() => this.setState({ preferencesOpen: false })}
        />
      </MaterialDrawer>
    );
  }
}

export default compose(withStyles(styles), withApollo)(Drawer);
