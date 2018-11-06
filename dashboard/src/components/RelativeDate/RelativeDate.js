import React from "react";
import PropTypes from "prop-types";
import capitalizeStr from "lodash/capitalize";
import Tooltip from "@material-ui/core/Tooltip";

class RelativeDate extends React.Component {
  static propTypes = {
    capitalize: PropTypes.bool,
    // Allows consumer to specify whether or not the delta should be clamped.
    //   "only-pending" equivelant to clamp(delta, 0, math.Infinity)
    //   "only-past" equivelant to clamp(delta, -math.Infinity, 0)
    //   the default option "none" means delta will not be clamped
    clamp: PropTypes.oneOf(["only-pending", "only-past", "none"]),
    dateTime: PropTypes.string.isRequired,
    precision: PropTypes.oneOf(["default", "seconds"]),
    // TODO: intl-relativeformat is out of date w/ specification
    // style: PropTypes.oneOf(["long", "short", "narrow"]),
    style: PropTypes.oneOf(["best fit", "numeric"]),
    to: PropTypes.instanceOf(Date).isRequired,
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
    clamp: "none",
    precision: "default",
    // TODO: intl-relativeformat is out of date w/ specification
    // style: "long",
    style: "best fit",
    unit: undefined,
  };

  static getDerivedStateFromProps(props, state) {
    let timestamp = new Date(props.dateTime).getTime();
    timestamp -= timestamp % (props.precision === "seconds" ? 1000 : 60000);

    if (state.timestamp !== timestamp) {
      return { timestamp };
    }
    return null;
  }

  state = {
    timestamp: null,
  };

  shouldComponentUpdate(nextProps, nextState) {
    return (
      nextState.timestamp !== this.state.timestamp ||
      this.props.to.valueOf() !== nextProps.to.valueOf()
    );
  }

  render() {
    const {
      capitalize,
      clamp,
      dateTime,
      precision,
      style,
      to,
      unit,
      ...props
    } = this.props;

    const formatter = new IntlRelativeFormat("en", { style });
    const dateValue = new Date(dateTime);

    let delta = to - dateValue;
    if (clamp === "only-past") {
      delta = Math.max(delta, 0);
    } else if (clamp === "only-pending") {
      delta = Math.min(delta, 0);
    }

    let relativeDate;
    if (Math.abs(delta) >= 60000 || precision === "seconds") {
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
