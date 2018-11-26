import React from "react";
import PropTypes from "prop-types";
import { Transition, animated } from "react-spring";
import ResizeObserver from "react-resize-observer";
import { withStyles } from "@material-ui/core/styles";

import { Well } from "/components/relocation/Relocation";
import { TOAST } from "/components/relocation/types";

import UnmountObserver from "/components/util/UnmountObserver";

const styles = theme => ({
  toast: {
    position: "relative",
    left: 0,
    right: 0,
  },
  toastPadding: {
    [theme.breakpoints.up("md")]: {
      paddingBottom: 10,
      paddingRight: 10,
    },
  },
});

class ToastWell extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
  };

  state = {
    heights: {},
  };

  handleToastSize = (id, rect) => {
    this.setState(state => {
      if (rect.height === state.heights[id]) {
        return null;
      }

      const heights = { ...state.heights, [id]: rect.height };

      return { heights };
    });
  };

  handleToastUnmount = id => {
    this.setState(state => {
      if (state.heights[id] === undefined) {
        return null;
      }

      const heights = {};

      Object.keys(state.heights).forEach(key => {
        if (key !== id) {
          heights[key] = state.heights[key];
        }
      });

      return { heights };
    });
  };

  render() {
    const { classes } = this.props;
    const { heights } = this.state;

    return (
      <Well>
        {({ elements }) => {
          const visibleElements = elements
            .filter(({ props }) => props.type === TOAST)
            .slice(-20);

          return (
            <Transition
              native
              keys={visibleElements.map(element => element.id)}
              from={{ opacity: 1, height: 0 }}
              update={id => ({ opacity: 1, height: heights[id] || 0 })}
              leave={{ opacity: 0, height: 0 }}
              config={{ tension: 210, friction: 20 }}
            >
              {visibleElements.map(({ id, props, remove }) => style => (
                <animated.div style={style} className={classes.toast}>
                  <div style={{ position: "relative" }}>
                    <ResizeObserver
                      onResize={rect => this.handleToastSize(id, rect)}
                    />
                    <UnmountObserver
                      onUnmount={() => this.handleToastUnmount(id)}
                    />
                    <div className={classes.toastPadding}>
                      {props.render({ id, remove })}
                    </div>
                  </div>
                </animated.div>
              ))}
            </Transition>
          );
        }}
      </Well>
    );
  }
}

export default withStyles(styles)(ToastWell);
