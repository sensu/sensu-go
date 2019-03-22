import React from "react";
import PropTypes from "prop-types";

class Duration extends React.PureComponent {
  static propTypes = {
    duration: PropTypes.number.isRequired,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
  };

  static defaultProps = {
    component: "span",
  };

  numberFormatter = opts => new Intl.NumberFormat("en-US", opts);

  formatDuration = duration => {
    if (duration < 1000) {
      // ex. 892ms
      const formatter = this.numberFormatter({ maximumFractionDigits: 0 });
      return `${formatter.format(duration)}ms`;
    }
    if (duration < 10000) {
      // ex. 8.02s
      const formatter = this.numberFormatter({
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      });
      return `${formatter.format(duration / 1000)}s`;
    }
    if (duration < 60000) {
      // ex. 15.2s
      const formatter = this.numberFormatter({
        minimumFractionDigits: 1,
        maximumFractionDigits: 1,
      });
      return `${formatter.format(duration / 1000)}s`;
    }

    // ex. 150s
    const formatter = this.numberFormatter({
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
    return `${formatter.format(duration / 1000)}s`;
  };

  render() {
    const { component: Component, duration, ...props } = this.props;
    const formattedDuration = this.formatDuration(duration);

    return <Component {...props}>{formattedDuration}</Component>;
  }
}

export default Duration;
