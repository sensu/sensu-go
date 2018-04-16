import React from "react";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  "@global": {
    html: {
      // Background
      background: theme.palette.background.default,

      // Prevent text from being selected unless explicitly overridden
      userSelect: "none",

      // Ensure text is antialiased
      WebkitFontSmoothing: "antialiased",
      MozOsxFontSmoothing: "grayscale",

      height: "100%",
    },

    body: {
      minHeight: "100%",
      display: "flex",
      flexDirection: "column",
    },

    "body > div.root": {
      display: "flex",
      flexDirection: "column",
      flex: 1,
      flexBasis: "auto",
    },

    button: {
      cursor: "pointer",
    },
  },
});

class ThemeStyles extends React.PureComponent {
  render() {
    return null;
  }
}

export default withStyles(styles)(ThemeStyles);
