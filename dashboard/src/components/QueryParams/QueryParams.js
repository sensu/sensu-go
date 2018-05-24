import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";

function expandParams(params, keys) {
  return Array.from(params).reduce((acc, entry) => {
    const [key, val] = entry;
    if (keys && keys.indexOf(key) === -1) return acc;

    const prevVal = acc[key];
    if (Array.isArray(prevVal)) {
      acc[key] = [val, ...prevVal];
    } else if (prevVal) {
      acc[key] = [val, prevVal];
    } else {
      acc[key] = val;
    }

    return acc;
  }, {});
}

class QueryParams extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    location: PropTypes.object.isRequired,
    history: PropTypes.object.isRequired,
    keys: PropTypes.arrayOf(PropTypes.string),
  };

  static defaultProps = {
    keys: null,
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
    const params = expandParams(this.params, this.props.keys);
    return this.props.children(params, this.changeQuery);
  }
}

export default withRouter(QueryParams);
