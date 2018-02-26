import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  content: theme.mixins.gutters({
    paddingTop: 100, // TODO: make non-magic number
    flex: "1 1 100%",
    maxWidth: "100%",
    margin: "0 auto",
    position: "relative",
  }),

  // TODO: Make extra-wide gutter optional. (Maybe configured w/ prop?)
  [theme.breakpoints.up("sm")]: {
    content: {
      paddingLeft: theme.spacing.unit * 3 + 40,
      paddingRight: theme.spacing.unit * 3 + 40,
    },
  },
  [theme.breakpoints.up(900 + theme.spacing.unit * 6)]: {
    content: {
      maxWidth: 1000,
    },
  },
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
