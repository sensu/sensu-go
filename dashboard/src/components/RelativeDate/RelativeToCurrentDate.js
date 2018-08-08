import React from "react";
import PropTypes from "prop-types";
import CurrentDateProvider from "/components/CurrentDateProvider";
import RelativeDate from "./RelativeDate";

class RelativeToCurrentDate extends React.PureComponent {
  static propTypes = {
    ...RelativeDate.propTypes,
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
