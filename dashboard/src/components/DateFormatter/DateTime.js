import React from "react";
import PropTypes from "prop-types";
import DateFormatter from "./DateFormatter";

class DateTime extends React.PureComponent {
  static propTypes = {
    dateTime: PropTypes.string.isRequired,
    short: PropTypes.bool,
  };

  static defaultProps = {
    short: false,
  };

  render() {
    const { dateTime, short, ...props } = this.props;
    if (short) {
      return (
        <DateFormatter
          month="short"
          day="numeric"
          hour="numeric"
          minute="numeric"
          dateTime={dateTime}
          {...props}
        />
      );
    }

    return (
      <DateFormatter
        year="numeric"
        weekday="narrow"
        month="narrow"
        day="numeric"
        hour="numeric"
        minute="numeric"
        dateTime={dateTime}
        {...props}
      />
    );
  }
}

export default DateTime;
