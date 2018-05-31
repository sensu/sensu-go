import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";
import QueryParams from "./QueryParams";

class QueryParam extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    name: PropTypes.string.isRequired,
  };

  render() {
    const { name, children } = this.props;
    return (
      <QueryParams keys={[name]}>
        {(params, set) => children(params.get(name), val => set(name, val))}
      </QueryParams>
    );
  }
}

export default withRouter(QueryParam);
