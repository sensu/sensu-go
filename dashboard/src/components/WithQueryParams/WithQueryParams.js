import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";

class WithQueryParams extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    location: PropTypes.object.isRequired,
    history: PropTypes.object.isRequired,
  };

  constructor(props) {
    super(props);
    this.params = new URLSearchParams(props.location.search);
  }

  componentWillReceiveProps(nextProps) {
    this.params = new URLSearchParams(nextProps.location.search);
  }

  shouldComponentUpdate(nextProps) {
    if (this.props.children !== nextProps.children) {
      return true;
    } else if (this.props.location.pathname !== nextProps.location.pathname) {
      return true;
    } else if (this.props.location.search !== nextProps.location.search) {
      return true;
    }
    return false;
  }

  changeQuery = (key, val) => {
    this.params.set(key, val);
    this.props.history.push(
      `${this.props.location.pathname}?${this.params.toString()}`,
    );
  };

  render() {
    return this.props.children(this.params, this.changeQuery);
  }
}

export default withRouter(WithQueryParams);
