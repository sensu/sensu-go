import React from "react";
import PropTypes from "prop-types";

const LOCALE = "en-US";

class DateFormatter extends React.PureComponent {
  static propTypes = {
    dateTime: PropTypes.string.isRequired,
    hour12: PropTypes.bool,
    weekday: PropTypes.string,
    era: PropTypes.string,
    year: PropTypes.string,
    month: PropTypes.string,
    day: PropTypes.string,
    hour: PropTypes.string,
    minute: PropTypes.string,
    second: PropTypes.string,
  };

  static defaultProps = {
    hour12: true,
    weekday: undefined,
    era: undefined,
    year: undefined,
    month: undefined,
    day: undefined,
    hour: undefined,
    minute: undefined,
    second: undefined,
  };

  render() {
    const {
      dateTime,
      weekday,
      era,
      year,
      month,
      day,
      hour,
      minute,
      second,
      hour12,
      ...props
    } = this.props;

    const date = new Date(dateTime);
    const formatter = Intl.DateTimeFormat(LOCALE, {
      weekday,
      era,
      year,
      month,
      day,
      hour,
      minute,
      second,
      hour12,
    });

    return (
      <time dateTime={dateTime} {...props}>
        {formatter.format(date)}
      </time>
    );
  }
}

export default DateFormatter;
