import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    height: 2,
    flexShrink: 0,
    border: "none",
    margin: "0 0 -2px 0",
  },
  primary: {
    backgroundColor: theme.palette.primary,
  },
  secondary: {
    backgroundColor: theme.palette.secondary,
  },
  success: {
    backgroundColor: theme.palette.success,
  },
  warning: {
    backgroundColor: theme.palette.warning,
  },
  critical: {
    backgroundColor: theme.palette.critical,
  },
  unknown: {
    backgroundColor: theme.palette.unknown,
  },
});

class CardHighlight extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    color: PropTypes.oneOf([
      "primary",
      "secondary",
      "success",
      "warning",
      "critical",
      "unknown",
    ]).isRequired,
  };

  render() {
    const { classes, color } = this.props;
    const className = classnames(classes.root, classes[color]);

    return <hr className={className} />;
  }
}

export default withStyles(styles)(CardHighlight);
