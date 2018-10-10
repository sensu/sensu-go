import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";

function expandParams(params, keys, defaults) {
  const matched = Array.from(params).reduce((acc, entry) => {
    const [key, val] = entry;
    if (keys && keys.indexOf(key) === -1) {
      return acc;
    }

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

  return keys.reduce((acc, key) => {
    if (!acc[key]) {
      acc[key] = defaults[key];
    }
    return acc;
  }, matched);
}

class QueryParams extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    defaults: PropTypes.object,
    history: PropTypes.object.isRequired,
    keys: PropTypes.arrayOf(PropTypes.string),
    location: PropTypes.object.isRequired,
  };

  static defaultProps = {
    keys: null,
    defaults: {},
  };

  static getDerivedStateFromProps(nextProps) {
    return {
      params: new URLSearchParams(nextProps.location.search),
    };
  }

  state = {};

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

  changeQuery = fnOrObj => {
    const params = Object.assign(this.state.params, {});
    params.reset = keys =>
      Array.from(keys || params.keys()).forEach(key => params.delete(key));

    if (typeof fnOrObj === "function") {
      fnOrObj(params);
    } else {
      Object.keys(fnOrObj).forEach(key => params.set(key, fnOrObj[key]));
    }

    const newPath = `${this.props.location.pathname}?${params.toString()}`;
    this.props.history.push(newPath);
  };

  render() {
    const params = expandParams(
      this.state.params,
      this.props.keys,
      this.props.defaults,
    );

    return this.props.children(params, this.changeQuery);
  }
}

export default withRouter(QueryParams);
