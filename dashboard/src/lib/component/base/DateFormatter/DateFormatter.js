import React from "react";
import PropTypes from "prop-types";

const LOCALE = "en-US";

class DateFormatter extends React.PureComponent {
  static propTypes = {
    value: PropTypes.instanceOf(Date).isRequired,
    hour12: PropTypes.bool,
    hourCycle: PropTypes.oneOf(["h11", "h12", "h23", "h24"]),
    weekday: PropTypes.oneOf(["narrow", "short", "long"]),
    era: PropTypes.oneOf(["narrow", "short", "long"]),
    year: PropTypes.oneOf(["numeric", "2-digit"]),
    month: PropTypes.oneOf(["numeric", "2-digit", "narrow", "short", "long"]),
    day: PropTypes.oneOf(["numeric", "2-digit"]),
    hour: PropTypes.oneOf(["numeric", "2-digit"]),
    minute: PropTypes.oneOf(["numeric", "2-digit"]),
    second: PropTypes.oneOf(["numeric", "2-digit"]),
    timeZoneName: PropTypes.oneOf(["short", "long"]),
  };

  static defaultProps = {
    hour12: true,
    hourCycle: undefined,
    weekday: undefined,
    era: undefined,
    year: undefined,
    month: undefined,
    day: undefined,
    hour: undefined,
    minute: undefined,
    second: undefined,
    timeZoneName: undefined,
  };

  render() {
    const {
      value,
      weekday,
      era,
      year,
      month,
      day,
      hour,
      minute,
      second,
      hour12,
      hourCycle,
      timeZoneName,
      ...props
    } = this.props;

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
      hourCycle,
      timeZoneName,
    });

    return (
      <time dateTime={value.toString()} {...props}>
        {formatter.format(value)}
      </time>
    );
  }
}

export default DateFormatter;
