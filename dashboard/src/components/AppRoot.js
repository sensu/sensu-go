import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

class AppRoot extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    children: PropTypes.node,
  };

  static defaultProps = { children: null };

  static styles = {
    root: {
      display: "flex",
      alignItems: "stretch",
      minHeight: "100vh",
      width: "100%",
    },
  };

  render() {
    const { children, classes } = this.props;
    return <div className={classes.root}>{children}</div>;
  }
}

const EnhancedAppRoot = withStyles(AppRoot.styles)(AppRoot);
export default EnhancedAppRoot;
