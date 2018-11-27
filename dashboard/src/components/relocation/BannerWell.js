import React from "react";
import PropTypes from "prop-types";
import { Transition, animated } from "react-spring";
import ResizeObserver from "react-resize-observer";
import { withStyles } from "@material-ui/core/styles";

import { Well } from "/components/relocation/Relocation";
import { BANNER } from "/components/relocation/types";

import UnmountObserver from "/components/util/UnmountObserver";

const MAX_BANNERS = 20;

const zIndices = {};

// Include additional styles to account for transitioning elements that may
// increase the visible count beyond the max.
for (let i = 0; i < MAX_BANNERS + 5; i += 1) {
  zIndices[`&:nth-child(${i})`] = { zIndex: MAX_BANNERS + 5 - i };
}

const styles = () => ({
  container: {
    position: "relative",
    marginTop: -8,
    top: 8,
    paddingBottom: 8,
    overflow: "hidden",
  },
  banner: {
    position: "relative",
    ...zIndices,
  },
  bannerInner: {
    position: "absolute",
    left: 0,
    right: 0,
    bottom: 0,
  },
});

class BannerWell extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
  };

  state = {
    heights: {},
  };

  handleBannerSize = (id, rect) => {
    this.setState(state => {
      if (rect.height === state.heights[id]) {
        return null;
      }

      const heights = { ...state.heights, [id]: rect.height };

      return { heights };
    });
  };

  handleBannerUnmount = id => {
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
            .filter(({ props }) => props.type === BANNER)
            .slice(-MAX_BANNERS);

          visibleElements.reverse();

          return (
            <div className={classes.container}>
              <Transition
                native
                keys={visibleElements.map(element => element.id)}
                from={{ opacity: 0, height: 0 }}
                update={id => ({ opacity: 1, height: heights[id] || 0 })}
                leave={{ opacity: 0, height: 0 }}
                config={{ tension: 210, friction: 20 }}
              >
                {visibleElements.map(({ id, props, remove }) => style => (
                  <animated.div style={style} className={classes.banner}>
                    <div className={classes.bannerInner}>
                      <ResizeObserver
                        onResize={rect => this.handleBannerSize(id, rect)}
                      />
                      <UnmountObserver
                        onUnmount={() => this.handleBannerUnmount(id)}
                      />
                      {props.render({ id, remove })}
                    </div>
                  </animated.div>
                ))}
              </Transition>
            </div>
          );
        }}
      </Well>
    );
  }
}

export default withStyles(styles)(BannerWell);
