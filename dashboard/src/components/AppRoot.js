import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

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
});

class AppRoot extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    children: PropTypes.element,
  };

  static defaultProps = { children: null };

  render() {
    const { children, classes } = this.props;
    return <div className={classes.root}>{children}</div>;
  }
}

export default withStyles(styles)(AppRoot);
