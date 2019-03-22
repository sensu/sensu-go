import React from "react";
import PropTypes from "prop-types";
import ResizeObserver from "react-resize-observer";

class CircularProgress extends React.PureComponent {
  static propTypes = {
    width: PropTypes.number,
    value: PropTypes.number,
    children: PropTypes.node,
    opacity: PropTypes.number,
  };

  static defaultProps = {
    width: 8,
    value: 0,
    children: undefined,
    opacity: 1,
  };

  state = {
    size: 0,
  };

  handleResize = rect => {
    this.setState(state => {
      const size = Math.min(rect.width, rect.height);
      if (size === state.size) {
        return null;
      }
      return { size };
    });
  };

  render() {
    const { size } = this.state;
    const { width, value, children, opacity } = this.props;

    return (
      <div style={{ position: "relative" }}>
        <ResizeObserver onResize={this.handleResize} />
        <svg
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block", position: "absolute" }}
        >
          {size > 0 && (
            <circle
              transform={`rotate(-90, ${size * 0.5}, ${size * 0.5})`}
              cx={size * 0.5}
              cy={size * 0.5}
              r={(size - width) / 2}
              strokeDasharray={Math.PI * (size - width)}
              strokeDashoffset={Math.PI * (size - width) * (1 - value)}
              fill="none"
              stroke="currentColor"
              opacity={opacity}
              strokeWidth={width}
            />
          )}
        </svg>
        {children}
      </div>
    );
  }
}

export default CircularProgress;
