import React from "react";
import PropTypes from "prop-types";
import capitalizeStr from "lodash/capitalize";
import Tooltip from "@material-ui/core/Tooltip";

// time interval in which delta is recalculated
const recalcInterval = 60000;
const precision = 60000;

class RelativeDate extends React.Component {
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

  static getDerivedStateFromProps(props, state) {
    let timestamp = new Date(props.dateTime).getTime();
    timestamp -= timestamp % precision;

    if (state.timestamp !== timestamp) {
      return { timestamp };
    }
    return null;
  }

  state = {
    timestamp: null,
    now: Date.now(),
  };

  componentDidMount() {
    this.interval = setInterval(this.setDelta, recalcInterval);
  }

  shouldComponentUpdate(nextProps, nextState) {
    return nextState.timestamp !== this.state.timestamp;
  }

  componentWillUnmount() {
    clearInterval(this.interval);
  }

  setDelta = () => {
    this.setState({ now: Date.now() });
  };

  render() {
    const { dateTime, capitalize, unit, style, ...props } = this.props;
    const { now } = this.state;

    const formatter = new IntlRelativeFormat("en", { style });
    const dateValue = new Date(dateTime);
    const delta = now - dateValue;

    let relativeDate;
    if (Math.abs(delta) >= 60000) {
      relativeDate = formatter.format(dateValue, unit);
    } else if (delta >= 10000) {
      relativeDate = "seconds ago";
    } else if (delta >= 0) {
      relativeDate = "just now";
    } else if (delta >= -10000) {
      relativeDate = "in a few seconds";
    } else if (delta >= -60000) {
      relativeDate = "in less than a minute";
    }
    if (capitalize) {
      relativeDate = capitalizeStr(relativeDate);
    }
    return (
      <Tooltip title={dateValue.toString()}>
        <time dateTime={dateTime} {...props}>
          {relativeDate}
        </time>
      </Tooltip>
    );
  }
}

export default RelativeDate;
