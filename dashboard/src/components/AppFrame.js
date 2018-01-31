import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import QuickNav from "./QuickNav";
import Drawer from "./Drawer";
import Toolbar from "./Toolbar";

const styles = theme => ({
  "@global": {
    html: {
      background: theme.palette.background.default,
      WebkitFontSmoothing: "antialiased", // Antialiasing.
      MozOsxFontSmoothing: "grayscale", // Antialiasing.
      boxSizing: "border-box",
    },
    "*, *:before, *:after": {
      boxSizing: "inherit",
    },
    body: {
      margin: 0,
    },
  },
  root: {
    display: "flex",
    alignItems: "stretch",
    minHeight: "100vh",
    width: "100%",
  },
  drawer: {
    [theme.breakpoints.up("lg")]: {
      width: 250,
    },
  },
  quicknav: {
    position: "fixed",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    top: 80,
    left: 0,
    width: 72,
  },
  maincontainer: {
    position: "relative",
    display: "flex",
    width: "100%",
  },
});

class AppFrame extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  state = {
    toolbar: false,
  };

  render() {
    const { children, classes } = this.props;
    const { toolbar } = this.state;

    const toggleToolbar = () => {
      this.setState({ toolbar: !toolbar });
    };

    return (
      <div className={classes.root}>
        <Toolbar toggleToolbar={toggleToolbar} />
        <Drawer
          open={toolbar}
          onToggle={toggleToolbar}
          className={classes.drawer}
        />
        <div className={classes.maincontainer}>
          <QuickNav className={classes.quicknav} />
          {children}
        </div>
      </div>
    );
  }
}
export default withStyles(styles)(AppFrame);
