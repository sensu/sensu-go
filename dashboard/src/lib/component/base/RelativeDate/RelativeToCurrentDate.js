import React from "react";
import PropTypes from "prop-types";

import CurrentDateProvider from "/lib/component/util/CurrentDateProvider";

import RelativeDate from "./RelativeDate";

class RelativeToCurrentDate extends React.PureComponent {
  static propTypes = {
    ...RelativeDate.propTypes,
    to: PropTypes.any, // eslint-disable-line react/require-default-props
    refreshInterval: PropTypes.number,
  };

  static defaultProps = {
    refreshInterval: 10000,
  };

  render() {
    const { refreshInterval, ...props } = this.props;

    return (
      <CurrentDateProvider interval={refreshInterval}>
        {now => <RelativeDate to={now} {...props} />}
      </CurrentDateProvider>
    );
  }
}

export default RelativeToCurrentDate;
