import React from "react";
import PropTypes from "prop-types";
import DateFormatter from "./DateFormatter";

class DateTime extends React.PureComponent {
  static propTypes = {
    value: PropTypes.instanceOf(Date).isRequired,
    short: PropTypes.bool,
  };

  static defaultProps = {
    short: false,
  };

  render() {
    const { value, short, ...props } = this.props;

    if (short) {
      return (
        <DateFormatter
          month="short"
          day="numeric"
          hour="numeric"
          minute="numeric"
          value={value}
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
        value={value}
        {...props}
      />
    );
  }
}

export default DateTime;
