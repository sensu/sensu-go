import React from "react";
import PropTypes from "prop-types";
import DateFormatter from "./DateFormatter";

class KitchenTime extends React.PureComponent {
  static propTypes = {
    value: PropTypes.instanceOf(Date).isRequired,
  };

  render() {
    const { value, ...props } = this.props;
    return (
      <DateFormatter value={value} hour="numeric" minute="numeric" {...props} />
    );
  }
}

export default KitchenTime;
