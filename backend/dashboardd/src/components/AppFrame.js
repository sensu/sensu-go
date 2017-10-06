import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

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
});

class AppFrame extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    children: PropTypes.element.isRequired,
  };

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
        {children}
      </div>
    );
  }
}

export default withStyles(styles)(AppFrame);
