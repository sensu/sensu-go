import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  root: {
    flex: "1 1 100%",
    maxWidth: "100%",
    margin: "0 auto",

    paddingTop: theme.spacing.unit * 2,
    [theme.breakpoints.up("md")]: {
      paddingTop: 0,
    },

    // remove gutters when least screen real estate is available, giving content
    // ability- if need be- to reach the edge of the screen.
    paddingLeft: 0,
    paddingRight: 0,

    // keep content from spanning too much of the screen and becoming difficult
    // to parse.
    [theme.breakpoints.up("sm")]: {
      maxWidth: 1080,
      paddingLeft: theme.spacing.unit,
      paddingRight: theme.spacing.unit,
    },
  },
  gutters: theme.mixins.gutters({
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
  }),
  [theme.breakpoints.up("lg")]: {
    fullWidth: {
      maxWidth: "initial",
    },
  },
});

class AppContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    fullWidth: PropTypes.bool,
    gutters: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    fullWidth: false,
    gutters: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      children,
      fullWidth,
      gutters,
    } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.fullWidth]: fullWidth,
      [classes.gutters]: gutters,
    });
    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(AppContent);
