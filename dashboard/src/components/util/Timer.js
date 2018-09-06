import React from "react";
import PropTypes from "prop-types";

class Timer extends React.PureComponent {
  static propTypes = {
    onEnd: PropTypes.func,
    delay: PropTypes.number.isRequired,
    children: PropTypes.func,
    paused: PropTypes.bool,
  };

  static defaultProps = {
    children: undefined,
    onEnd: undefined,
    paused: false,
  };

  static getDerivedStateFromProps(props, state) {
    if (state.pausedTime !== null && !props.paused) {
      const timePaused = Date.now() - state.pausedTime;

      return {
        pausedTime: null,
        startTime: state.startTime + timePaused,
      };
    }

    if (state.pausedTime === null && props.paused) {
      return {
        pausedTime: state.currentTime,
      };
    }

    return null;
  }

  constructor(props) {
    super(props);

    const now = Date.now();

    this.state = {
      currentTime: now,
      startTime: now,
      pausedTime: null,
    };
  }

  componentDidMount() {
    if (!this.props.paused) {
      this.animationFrameRef = requestAnimationFrame(this.tick);
    }
  }

  componentDidUpdate() {
    const { startTime, currentTime } = this.state;
    const { delay, onEnd, paused } = this.props;

    if (currentTime >= startTime + delay) {
      if (onEnd && !this.onEndCalled) {
        this.onEndCalled = true;
        onEnd();
      }
    } else if (!this.animationFrameRef && !paused) {
      this.animationFrameRef = requestAnimationFrame(this.tick);
    }
  }

  componentWillUnmount() {
    cancelAnimationFrame(this.animationFrameRef);
    this.animationFrameRef = null;
    this.wilUnmount = true;
  }

  tick = () => {
    this.animationFrameRef = null;
    this.setState({ currentTime: Date.now() });
  };

  render() {
    const { startTime, currentTime } = this.state;
    const { children, delay } = this.props;
    return children
      ? children(Math.min(1, (currentTime - startTime) / delay))
      : null;
  }
}

export default Timer;
