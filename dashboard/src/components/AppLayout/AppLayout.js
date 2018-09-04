import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import ResizeObserver from "react-resize-observer";

import MobileFullWidthContent from "./MobileFullWidthContent";
import Context from "./Context";

const styles = theme => ({
  root: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    alignItems: "stretch",
  },

  topBarContainer: {
    flex: 0,
    position: "fixed",
    top: 0,
    left: 0,
    right: 0,

    zIndex: 1,

    "@supports (position: sticky)": {
      position: "sticky",
    },
  },

  topBarObserver: {
    position: "relative",
  },

  quickNavContainer: {
    position: "relative",
  },

  quickNav: {
    position: "absolute",
    flexDirection: "column",
    alignItems: "stretch",
    width: 72,
    display: "none",

    paddingTop: 12,

    [theme.breakpoints.up("md")]: {
      display: "flex",
    },
  },

  alertContainer: {
    position: "relative",
  },

  alert: {
    position: "absolute",
    left: 0,
    right: 0,
    height: 0,
  },

  contentContainer: {
    flex: 1,
    display: "flex",
  },

  content: {
    flex: 1,
    position: "relative",
    display: "flex",
    flexDirection: "column",
    alignItems: "stretch",

    marginLeft: "auto",
    marginRight: "auto",

    maxWidth: 1224,

    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,

    paddingTop: 24,
    paddingBottom: 24,

    [theme.breakpoints.up("md")]: {
      // add gutters for quick nav and any floating actions.
      paddingLeft: 80,
      paddingRight: 80,
    },
  },

  toastContainer: {
    position: "fixed",
    bottom: 0,
    left: 0,
  },

  toast: {
    position: "absolute",
    bottom: 0,
    left: 0,
  },
});

class AppLayout extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    topBar: PropTypes.node,
    quickNav: PropTypes.node,
    content: PropTypes.node,
    alert: PropTypes.node,
    toast: PropTypes.node,
  };

  static defaultProps = {
    topBar: undefined,
    quickNav: undefined,
    content: undefined,
    alert: undefined,
    toast: undefined,
  };

  static MobileFullWidthContent = MobileFullWidthContent;
  static Context = Context;

  state = { topBarHeight: 0 };

  handleTopBarResize = rect => {
    this.setState(state => {
      if (state.topBarHeight === rect.height) {
        return null;
      }

      return { topBarHeight: rect.height };
    });
  };

  render() {
    const { classes, topBar, quickNav, content, alert, toast } = this.props;

    const contentOffset =
      CSS && CSS.supports && CSS.supports("position: sticky")
        ? 0
        : this.state.topBarHeight;

    return (
      <Context.Provider value={this.state}>
        <div className={classes.root}>
          <div className={classes.topBarContainer}>
            <div className={classes.topBarObserver}>
              <ResizeObserver onResize={this.handleTopBarResize} />
              {topBar}
            </div>
            <div className={classes.quickNavContainer}>
              <div className={classes.quickNav}>{quickNav}</div>
            </div>
            <div className={classes.alertContainer}>
              <div className={classes.alert}>{alert}</div>
            </div>
          </div>
          <div style={{ height: contentOffset }} />
          <div className={classes.contentContainer}>
            <div className={classes.content}>{content}</div>
          </div>
          <div className={classes.toastContainer}>
            <div className={classes.toast}>{toast}</div>
          </div>
        </div>
      </Context.Provider>
    );
  }
}

export default withStyles(styles)(AppLayout);
