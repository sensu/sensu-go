import React from "react";
import PropTypes from "prop-types";

class Timer extends React.PureComponent {
  static propTypes = {
    onEnd: PropTypes.func,
    delay: PropTypes.number.isRequired,
    children: PropTypes.func,
  };

  static defaultProps = {
    children: undefined,
    onEnd: undefined,
  };

  constructor(props) {
    super(props);

    const now = Date.now();

    this.state = {
      currentTime: now,
      startTime: now,
    };
  }

  componentDidMount() {
    this.animationFrameRef = requestAnimationFrame(this.tick);
  }

  componentDidUpdate() {
    const { startTime, currentTime } = this.state;
    const { delay, onEnd } = this.props;

    if (currentTime >= startTime + delay) {
      if (onEnd && !this.onEndCalled) {
        this.onEndCalled = true;
        onEnd();
      }
    } else if (!this.animationFrameRef) {
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
