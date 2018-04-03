import React from "react";
import { withStyles } from "material-ui/styles";

// Largely borrows from sanatize.css and normalize.css
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

      // Add border box sizing in all browsers
      boxSizing: "border-box",

      // Add the default cursor in all browsers
      cursor: "default",

      // Prevent font size adjustments after orientation changes in IE and iOS.
      "-webkit-text-size-adjust": "100%",
      "ms-text-size-adjust": "100%",
    },

    "*, *:before, *:after": {
      // Add border box sizing in all browsers
      boxSizing: "inherit",

      // Remove repeating backgrounds in all browsers
      backgroundRepeat: "no-repeat",

      // Add text decoration inheritance in all browsers
      textDecoration: "inherit",

      // Add vertical alignment inheritence in all browsers
      verticalAlign: "inherit",
    },

    hr: {
      // Correct box-sizing in FF
      boxSizing: "content-box",

      // Show overflow in IE.
      height: 0,
      overflow: "visible",
    },

    "nav ol, nav ul": {
      // Remove list style
      listStyle: "none",
    },

    a: {
      // Remove the gray background on active links in IE 10.
      backgroundColor: "transparent",

      // Remove gaps in links underline in iOS 8+ and Safari 8+.
      "webkit-text-decoration-skip": "objects",
    },

    img: {
      // Remove border style from linked images on IE.
      borderStyle: "none",
    },

    "svg:not(:root)": {
      // Hide the overflow in IE.
      overflow: "hidden",
    },

    body: {
      // Remove margin.
      margin: 0,
    },
  },
});

class BrowserOverrides extends React.PureComponent {
  render() {
    return null;
  }
}

const EnhancedBrowserOverrides = withStyles(styles)(BrowserOverrides);
export default EnhancedBrowserOverrides;
