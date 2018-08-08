import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    display: "flex",
    alignItems: "center",
  },
  gutters: {
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    [theme.breakpoints.up("sm")]: {
      paddingLeft: 0,
      paddingRight: 0,
    },
  },
  bottomMargin: {
    marginBottom: theme.spacing.unit * 2,
  },
});

class Content extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    container: PropTypes.bool,
    bottomMargin: PropTypes.bool,
    gutters: PropTypes.bool,
  };

  static defaultProps = {
    bottomMargin: false,
    className: "",
    container: false,
    gutters: false,
  };

  render() {
    const {
      children,
      classes,
      className: classNameProp,
      container,
      bottomMargin,
      gutters,
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.bottomMargin]: bottomMargin,
      [classes.gutters]: gutters,
    });

    if (container || React.Children.count(children) > 1) {
      return <div className={className}>{children}</div>;
    }
    return React.cloneElement(children, { className });
  }
}

export default withStyles(styles)(Content);
