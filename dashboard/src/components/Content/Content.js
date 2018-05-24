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
    bottomMargin: PropTypes.bool,
    gutters: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    bottomMargin: false,
    gutters: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      children,
      bottomMargin,
      gutters,
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.bottomMargin]: bottomMargin,
      [classes.gutters]: gutters,
    });
    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(Content);
