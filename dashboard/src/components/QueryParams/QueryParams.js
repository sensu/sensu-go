import React from "react";
import PropTypes from "prop-types";
import { withRouter } from "react-router-dom";

function expandParams(params, keys) {
  return Array.from(params).reduce((acc, entry) => {
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

    if (typeof fnOrObj === "function") {
      fnOrObj(params);
    } else {
      Object.keys(fnOrObj).forEach(key => params.set(key, fnOrObj[key]));
    }

    const newPath = `${this.props.location.pathname}?${params.toString()}`;
    this.props.history.push(newPath);
  };

  render() {
    const params = expandParams(this.state.params, this.props.keys);
    return this.props.children(params, this.changeQuery);
  }
}

export default withRouter(QueryParams);
