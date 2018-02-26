import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  content: theme.mixins.gutters({
    flex: "1 1 100%",
    maxWidth: "100%",
    margin: "0 auto",
    [theme.breakpoints.up("md")]: {
      paddingTop: theme.spacing.unit * 3,
    },
    [theme.breakpoints.up("lg")]: {
      maxWidth: 1080,
    },
  }),
});

class AppContent extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    children: PropTypes.element.isRequired,
  };

  render() {
    const { classes, children } = this.props;
    return <div className={classes.content}>{children}</div>;
  }
}

export default withStyles(styles)(AppContent);
