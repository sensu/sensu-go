import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
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
