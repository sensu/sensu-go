import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    display: "flex",
  },
  gutters: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
    [theme.breakpoints.up("md")]: {
      margin: 0,
    },
  },
  marginBottom: {
    marginBottom: theme.spacing.unit,
    [theme.breakpoints.up("xs")]: {
      marginBottom: theme.spacing.unit * 2,
    },
  },
});

class Content extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    marginBottom: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    marginBottom: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      children,
      marginBottom,
    } = this.props;

    const className = classnames(classes.root, classes.gutter, classNameProp, {
      [classes.marginBottom]: marginBottom,
    });
    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(Content);
