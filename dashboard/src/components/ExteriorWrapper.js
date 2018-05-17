import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import ThemeProvider from "/components/ThemeProvider";

class ExteriorWrapper extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
    classes: PropTypes.shape({ root: PropTypes.string }).isRequired,
  };

  static styles = theme => ({
    "@global": {
      html: {
        background: theme.palette.primary["500"],
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

  render() {
    const { classes, children } = this.props;
    return (
      <ThemeProvider>
        <div className={classes.root}>{children}</div>
      </ThemeProvider>
    );
  }
}

export default withStyles(ExteriorWrapper.styles)(ExteriorWrapper);
