import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  content: theme.mixins.gutters({
    flex: "1 1 100%",
    maxWidth: "100%",
    margin: "0 auto",

    paddingTop: theme.spacing.unit * 2,
    [theme.breakpoints.up("md")]: {
      paddingTop: 0,
    },

    // remove gutters on mobile.
    paddingLeft: 0,
    paddingRight: 0,

    // keep content from spanning too much space and becoming difficult to
    // parse.
    [theme.breakpoints.up("lg")]: {
      maxWidth: 1080,
    },
  }),
});

class AppContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.element.isRequired,
  };

  static defaultProps = {
    className: "",
  };

  render() {
    const { classes, className, children } = this.props;
    const contentCls = classnames(classes.content, className);
    return <div className={contentCls}>{children}</div>;
  }
}

export default withStyles(styles)(AppContent);
