import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import gql from "graphql-tag";

import AppRoot from "./AppRoot";
import AppBar from "./AppBar";
import Drawer from "./Drawer";
import QuickNav from "./QuickNav";

class AppFrame extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object,
    environment: PropTypes.object,
    loaded: PropTypes.bool,
    children: PropTypes.element,
  };

  static defaultProps = {
    children: null,
    loaded: false,
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
    };
  };

  state = {
    drawerOpen: false,
  };

  render() {
    const { children, loaded, viewer, environment, classes } = this.props;
    const { drawerOpen } = this.state;

    const toggleDrawer = () => {
      this.setState({ drawerOpen: !drawerOpen });
    };

    return (
      <AppRoot>
        <AppBar
          loaded={loaded}
          environment={environment}
          toggleToolbar={toggleDrawer}
        />
        <Drawer
          loaded={loaded}
          viewer={viewer}
          open={drawerOpen}
          onToggle={toggleDrawer}
          environment={environment}
          className={classes.drawer}
        />
        <div className={classes.maincontainer}>
          <QuickNav className={classes.quicknav} />
          {children}
        </div>
      </AppRoot>
    );
  }
}

export default withStyles(AppFrame.styles)(AppFrame);
