import React from "react";
import PropTypes from "prop-types";
import warning from "warning";
import isNaN from "lodash/isNaN";
import DateFormatter from "./DateFormatter";

class DateStringFormatter extends React.PureComponent {
  static propTypes = {
    dateTime: PropTypes.string.isRequired,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
  };

  static defaultProps = {
    component: DateFormatter,
  };

  render() {
    const { dateTime, component: Component, ...props } = this.props;
    const ts = Date.parse(dateTime);

    if (isNaN(ts)) {
      warning("Unable to parse dateTime prop give to DateStringFormatter");
      return null;
    }

    const value = new Date(ts);
    return <Component {...props} value={value} />;
  }
}

export default DateStringFormatter;
