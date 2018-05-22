import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import gql from "graphql-tag";
import { Route } from "react-router-dom";

import AppBar from "/components/AppBar";
import Drawer from "/components/Drawer";
import QuickNav from "/components/QuickNav";
import Loader from "/components/util/Loader";

class AppFrame extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object,
    environment: PropTypes.object,
    loading: PropTypes.bool,
    children: PropTypes.element,
  };

  static defaultProps = {
    children: null,
    loading: false,
    viewer: null,
    environment: null,
  };

  static fragments = {
    viewer: gql`
      fragment AppFrame_viewer on Viewer {
        ...Drawer_viewer
      }

      ${Drawer.fragments.viewer}
    `,

    environment: gql`
      fragment AppFrame_environment on Environment {
        ...AppBar_environment
        ...Drawer_environment
      }

      ${AppBar.fragments.environment}
      ${Drawer.fragments.environment}
    `,
  };

  static styles = theme => {
    const toolbar = theme.mixins.toolbar;
    const xsBrk = `${theme.breakpoints.up("xs")} and (orientation: landscape)`;
    const smBrk = theme.breakpoints.up("sm");

    return {
      drawer: {
        [theme.breakpoints.up("lg")]: {
          width: 250,
        },
      },
      quicknav: {
        position: "fixed",
        flexDirection: "column",
        alignItems: "center",
        top: 80,
        left: 0,
        width: 72,
        display: "none",
        marginTop: "env(safe-area-inset-top)",
        [theme.breakpoints.up("md")]: {
          display: "flex",
        },
      },
      maincontainer: {
        position: "relative",
        display: "flex",
        width: "100%",
        flex: 1,
        marginTop: "env(safe-area-inset-top)",

        // Contend with app bar height.
        paddingTop: toolbar.minHeight,
        [xsBrk]: {
          paddingTop: toolbar[xsBrk].minHeight,
        },
        [smBrk]: {
          paddingTop: toolbar[smBrk].minHeight,
        },

        [theme.breakpoints.up("md")]: {
          // add gutters for quick nav and any floating actions.
          paddingLeft: 72,
          paddingRight: 72,

          // align content w/ top of quick nav
          paddingTop: toolbar[smBrk].minHeight + theme.spacing.unit * 3,
        },
      },

      appFrame: {
        minHeight: "100vh",
        width: "100%",
      },
    };
  };

  state = {
    drawerOpen: false,
  };

  render() {
    const { children, loading, viewer, environment, classes } = this.props;
    const { drawerOpen } = this.state;

    const toggleDrawer = () => {
      this.setState({ drawerOpen: !drawerOpen });
    };

    return (
      <Loader className={classes.appFrame} loading={loading}>
        <AppBar environment={environment} toggleToolbar={toggleDrawer} />
        <Drawer
          loading={loading}
          viewer={viewer}
          open={drawerOpen}
          onToggle={toggleDrawer}
          environment={environment}
          className={classes.drawer}
        />
        <div className={classes.maincontainer}>
          <Route
            path="/:organization/:environment"
            render={({ match: { params } }) => (
              <QuickNav
                className={classes.quicknav}
                organization={params.organization}
                environment={params.environment}
              />
            )}
          />
          {children}
        </div>
      </Loader>
    );
  }
}

export default withStyles(AppFrame.styles)(AppFrame);
