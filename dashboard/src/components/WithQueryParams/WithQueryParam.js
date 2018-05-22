import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";
import WithQueryParams from "./WithQueryParams";

class WithQueryParam extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    name: PropTypes.string.isRequired,
  };

  render() {
    const { name, children } = this.props;
    return (
      <WithQueryParams>
        {(params, set) => children(params.get(name), val => set(name, val))}
      </WithQueryParams>
    );
  }
}

export default withRouter(WithQueryParam);
