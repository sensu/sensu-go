import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import ResizeObserver from "react-resize-observer";

import ToastWell from "/lib/component/relocation/ToastWell";
import BannerWell from "/lib/component/relocation/BannerWell";

import MobileFullWidthContent from "./MobileFullWidthContent";
import Context from "./Context";

const styles = theme => ({
  root: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    alignItems: "stretch",
    paddingLeft: "env(safe-area-inset-left)",
    paddingRight: "env(safe-area-inset-right)",
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

  topBar: {
    position: "relative",
    zIndex: 1,
  },

  banner: {
    position: "relative",
    zIndex: 0,
  },

  contentContainer: {
    flex: 1,
    display: "flex",
    zIndex: 0,
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

    paddingTop: 16,
    paddingBottom: 24,

    [theme.breakpoints.up("md")]: {
      // align with quick nav container
      paddingTop: 24,

      // add gutters for quick nav and any floating actions.
      paddingLeft: 80,
      paddingRight: 80,
    },
  },

  toastContainer: {
    position: "fixed",
    bottom: 0,
    left: 0,
    right: 0,
    height: 0,
  },

  toast: {
    position: "absolute",
    bottom: 0,
    right: 0,
    left: 0,

    [theme.breakpoints.up("md")]: {
      left: "auto",
    },
  },
});

class AppLayout extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    topBar: PropTypes.node,
    quickNav: PropTypes.node,
    content: PropTypes.node,
  };

  static defaultProps = {
    topBar: undefined,
    quickNav: undefined,
    content: undefined,
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
    const { classes, topBar, quickNav, content } = this.props;

    const contentOffset =
      CSS && CSS.supports && CSS.supports("position: sticky")
        ? 0
        : this.state.topBarHeight;

    return (
      <Context.Provider value={this.state}>
        <div className={classes.root}>
          <div className={classes.topBarContainer}>
            <ResizeObserver onResize={this.handleTopBarResize} />
            <div className={classes.topBar}>{topBar}</div>
            <div className={classes.banner}>
              <BannerWell />
            </div>
            <div className={classes.quickNavContainer}>
              <div className={classes.quickNav}>{quickNav}</div>
            </div>
          </div>
          <div style={{ height: contentOffset }} />
          <div className={classes.contentContainer}>
            <div className={classes.content}>{content}</div>
          </div>
          <div className={classes.toastContainer}>
            <div className={classes.toast}>
              <ToastWell />
            </div>
          </div>
        </div>
      </Context.Provider>
    );
  }
}

export default withStyles(styles)(AppLayout);
