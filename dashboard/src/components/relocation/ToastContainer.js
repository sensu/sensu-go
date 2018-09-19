/* eslint-disable react/no-multi-comp */
import React from "react";
import PropTypes from "prop-types";
import ResizeObserver from "react-resize-observer";

import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    position: "relative",
    left: 0,
    right: 0,

    [theme.breakpoints.up("md")]: {
      left: "auto",
      right: 0,
    },
  },
  padding: {
    [theme.breakpoints.up("md")]: {
      paddingBottom: 10,
      paddingRight: 10,
    },
  },
});

class ToastContainer extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    onResize: PropTypes.func.isRequired,
    onUnmount: PropTypes.func.isRequired,
    classes: PropTypes.object.isRequired,
  };

  componentWillUnmount() {
    this.props.onUnmount();
  }

  render() {
    const { children, onResize, classes } = this.props;

    return (
      <div className={classes.root}>
        <ResizeObserver onResize={onResize} />
        <div className={classes.padding}>{children}</div>
      </div>
    );
  }
}

export default withStyles(styles)(ToastContainer);
