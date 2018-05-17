import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import ResizeObserver from "react-resize-observer";
import { Motion, spring, presets } from "react-motion";
import classnames from "classnames";

import AnimatedLogo from "/components/AnimatedLogo";

const LOADER_OPACITY = 1;
const OVERLAY_OPACITY = 0.8;

class Loader extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    loading: PropTypes.bool,
    children: PropTypes.any,
    delay: PropTypes.number,
    style: PropTypes.object,
    className: PropTypes.string,
    passthrough: PropTypes.bool,
  };

  static defaultProps = {
    loading: false,
    passthrough: false,
    children: null,
    delay: 100,
    className: undefined,
    style: {},
  };

  static styles = theme => ({
    container: {
      position: "relative",
      display: "flex",
      flexDirection: "column",
      flex: 1,
      flexBasis: "auto",
      alignItems: "stretch",
    },

    children: {
      zIndex: 0,
    },

    overlay: {
      zIndex: 1,
      position: "absolute",
      top: 0,
      right: 0,
      bottom: 0,
      left: 0,
      color: theme.palette.secondary.main,

      "&::before": {
        display: "block",
        content: '""',
        position: "absolute",
        top: 0,
        right: 0,
        bottom: 0,
        left: 0,
        background: theme.palette.background.paper,
        opacity: OVERLAY_OPACITY,
      },
    },

    loading: {
      "& $container $overlay": {
        display: "none",
      },
    },
  });

  state = {
    visible: false,
    spinnerPosition: { top: 0, left: 0, width: 0, height: 0 },
  };

  // eslint-disable-next-line react/sort-comp
  _timeout = null;

  componentWillMount() {
    if (this.props.loading) {
      this._setVisible(this.props.delay);
    }
  }

  componentWillReceiveProps(nextProps) {
    if (!this.props.loading && nextProps.loading) {
      this._setVisible(nextProps.delay);
    }

    if (this.props.loading && !nextProps.loading) {
      this._setHidden();
    }
  }

  _handleRect = rect => {
    const top = Math.max(0, -rect.top);
    const left = Math.max(0, -rect.left);
    const bottom = Math.max(0, rect.bottom - window.innerHeight);
    const right = Math.max(0, rect.right - window.innerWidth);

    this.setState({
      spinnerPosition: { top, left, bottom, right },
    });
  };

  _setVisible(delay) {
    if (!this._timeout) {
      this._timeout = setTimeout(() => this.setState({ visible: true }), delay);
    }
  }

  _setHidden() {
    if (this._timeout) {
      clearTimeout(this._timeout);
      this._timeout = null;
    }

    if (this.state.visible) {
      this.setState({ visible: false });
    }
  }

  componentWillUnmount() {
    clearTimeout(this._timeout);
  }

  render() {
    const { classes, passthrough, children, className, style } = this.props;
    const { visible, spinnerPosition } = this.state;

    const overlay = (
      <Motion style={{ progress: spring(visible ? 1 : 0, presets.noWobble) }}>
        {({ progress }) =>
          progress > 0.1 && (
            <React.Fragment>
              <ResizeObserver onReflow={this._handleRect} />
              <div className={classes.overlay} style={{ opacity: progress }}>
                <AnimatedLogo
                  size={50}
                  style={{
                    position: "absolute",
                    margin: "auto",
                    opacity: LOADER_OPACITY,
                    ...spinnerPosition,
                  }}
                />
              </div>
            </React.Fragment>
          )
        }
      </Motion>
    );

    if (passthrough) {
      return (
        <React.Fragment>
          {children}
          {overlay}
        </React.Fragment>
      );
    }

    return (
      <div
        className={classnames(className, classes.container, {
          [classes.loading]: visible,
        })}
        style={style}
      >
        <div className={classnames(classes.container, classes.children)}>
          {children}
        </div>
        {overlay}
      </div>
    );
  }
}

export default withStyles(Loader.styles)(Loader);
