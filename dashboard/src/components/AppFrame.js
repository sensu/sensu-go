import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import { createFragmentContainer, graphql } from "react-relay";

import AppRoot from "./AppRoot";
import AppBar from "./Toolbar";
import Drawer from "./Drawer";
import QuickNav from "./QuickNav";

const styles = theme => ({
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
    [theme.breakpoints.up("md")]: {
      display: "flex",
    },
  },
  maincontainer: {
    position: "relative",
    display: "flex",
    width: "100%",
    paddingTop: 64,
    [theme.breakpoints.up("md")]: {
      paddingLeft: 72,
      paddingRight: 72,
    },
  },
});

class AppFrame extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object.isRequired,
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  state = {
    drawerOpen: false,
  };

  render() {
    const { children, viewer, classes } = this.props;
    const { drawerOpen } = this.state;

    const toggleDrawer = () => {
      this.setState({ drawerOpen: !drawerOpen });
    };

    return (
      <AppRoot>
        <AppBar toggleToolbar={toggleDrawer} />
        <Drawer
          viewer={viewer}
          open={drawerOpen}
          onToggle={toggleDrawer}
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
export default createFragmentContainer(
  withStyles(styles)(AppFrame),
  graphql`
    fragment AppFrame_viewer on Viewer {
      ...Drawer_viewer
    }
  `,
);
