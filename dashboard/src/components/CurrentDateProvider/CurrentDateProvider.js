import React from "react";
import PropTypes from "prop-types";

class CurrentDateProvider extends React.PureComponent {
  static propTypes = {
    interval: PropTypes.number,
    children: PropTypes.func.isRequired,
  };

  static defaultProps = {
    interval: 1000,
  };

  state = {
    now: new Date(),
  };

  componentDidMount() {
    this.interval = setInterval(this.setDate, this.props.interval);
  }

  componentWillUnmount() {
    clearInterval(this.interval);
  }

  setDate = () => {
    this.setState({ now: new Date() });
  };

  render() {
    return this.props.children(this.state.now);
  }
}

export default CurrentDateProvider;
