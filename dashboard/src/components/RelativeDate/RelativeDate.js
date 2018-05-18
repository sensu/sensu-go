import React from "react";
import PropTypes from "prop-types";
import capitalizeStr from "lodash/capitalize";

// time interval in which delta is recalculated
const recalcInterval = 30000;

class RelativeDate extends React.PureComponent {
  static propTypes = {
    capitalize: PropTypes.bool,
    dateTime: PropTypes.string.isRequired,
    // TODO: intl-relativeformat is out of date w/ specification
    // style: PropTypes.oneOf(["long", "short", "narrow"]),
    style: PropTypes.oneOf(["best fit", "numeric"]),
    unit: PropTypes.oneOf([
      "year",
      "quarter",
      "month",
      "week",
      "day",
      "hour",
      "minute",
      "second",
    ]),
  };

  static defaultProps = {
    capitalize: false,
    // TODO: intl-relativeformat is out of date w/ specification
    // style: "long",
    style: "best fit",
    unit: undefined,
  };

  state = {
    now: null,
  };

  componentDidMount() {
    this.interval = setInterval(this.setDelta, recalcInterval);
  }

  componentWillUnmount() {
    clearInterval(this.interval);
  }

  setDelta = () => {
    this.setState({ now: Date.now() });
  };

  render() {
    const { dateTime, capitalize, style, unit, ...props } = this.props;
    const dateValue = new Date(dateTime);
    const formatter = new IntlRelativeFormat("en", { style });

    let relativeDate = formatter.format(dateValue, unit);
    if (capitalize) relativeDate = capitalizeStr(relativeDate);
    return (
      <time dateTime={dateTime} {...props}>
        {relativeDate}
      </time>
    );
  }
}

export default RelativeDate;
