import React from "react";
import PropTypes from "prop-types";
import { Route, Link } from "react-router-dom";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import { compose } from "recompose";

import MaterialDrawer from "@material-ui/core/Drawer";
import List from "@material-ui/core/List";
import Divider from "@material-ui/core/Divider";
import { withStyles } from "@material-ui/core/styles";
import FeedbackIcon from "@material-ui/icons/Feedback";
import LogoutIcon from "@material-ui/icons/ExitToApp";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";

import Loader from "/lib/component/base/Loader";

import Preferences from "/lib/component/base/Preferences";
import CheckIcon from "/lib/component/icon/Check";
import EntityIcon from "/lib/component/icon/Entity";
import EventIcon from "/lib/component/icon/Event";
import SilenceIcon from "/lib/component/icon/Silence";
import WandIcon from "/lib/component/icon/Wand";
import Wordmark from "/lib/component/icon/SensuWordmark";

import DrawerButton from "/app/component/partial/DrawerButton";
import NamespaceIcon from "/app/component/partial/NamespaceIcon";
import NamespaceSelector from "/app/component/partial/NamespaceSelector";

import invalidateTokens from "/lib/mutation/invalidateTokens";

const linkPath = ({ namespace }, path) => `/${namespace}/${path}`;

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
    height: 56,
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
    namespace: PropTypes.object,
    onToggle: PropTypes.func.isRequired,
    open: PropTypes.bool.isRequired,
    loading: PropTypes.bool,
  };

  static defaultProps = { loading: false, viewer: null, namespace: null };

  static fragments = {
    viewer: gql`
      fragment Drawer_viewer on Viewer {
        ...NamespaceSelector_viewer
      }

      ${NamespaceSelector.fragments.viewer}
      ${NamespaceSelector.fragments.namespace}
    `,

    namespace: gql`
      fragment Drawer_namespace on Namespace {
        ...NamespaceIcon_namespace
        ...NamespaceSelector_namespace
      }

      ${NamespaceSelector.fragments.namespace}
      ${NamespaceIcon.fragments.namespace}
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
      namespace,
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
                    {namespace && (
                      <NamespaceIcon namespace={namespace} size={36} />
                    )}
                  </div>
                </div>
                <div className={classes.row}>
                  <NamespaceSelector
                    viewer={viewer}
                    namespace={namespace}
                    className={classes.namespaceSelector}
                    onChange={onToggle}
                  />
                </div>
              </div>
            </div>
            <Divider />
            <Route
              path="/:namespace"
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
                  <DrawerButton
                    Icon={SilenceIcon}
                    primary="Silences"
                    component={Link}
                    onClick={onToggle}
                    to={linkPath(params, "silences")}
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

export default compose(
  withStyles(styles),
  withApollo,
)(Drawer);
