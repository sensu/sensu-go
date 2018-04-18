import React from "react";
import PropTypes from "prop-types";
import DateFormatter from "./DateFormatter";

class KitchenTime extends React.PureComponent {
  static propTypes = {
    dateTime: PropTypes.string.isRequired,
  };

  render() {
    const { dateTime, ...props } = this.props;
    return (
      <DateFormatter
        dateTime={dateTime}
        hour="numeric"
        minute="numeric"
        {...props}
      />
    );
  }
}

export default KitchenTime;
